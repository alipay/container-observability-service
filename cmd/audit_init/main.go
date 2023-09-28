package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v9"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

const (
	// container path for node root
	nodeRootPath = "/node/root"
	// apiserver pod manifest config
	apiserverManifestDir = "/etc/kubernetes/manifests"
	apiserverPodConfig   = "kube-apiserver.yaml"

	auditLogHostPath  = "/var/log/k8s-audit"
	auditLogMountPath = "/audit"
	auditLogFileName  = "audit.log"

	auditPolicyConfigMap = "apiserver-audit-config"
	auditPolicyFileName  = "audit-policy.yaml"
	auditPolicyHostPath  = "/etc/kubernetes/audit"
	auditPolicyMountPath = "/etc/kubernetes"

	// container flag
	// You can pass a file with the policy to kube-apiserver using the --audit-policy-file flag.
	// If the flag is omitted, no events are logged. Note that the rules field must be provided in the audit policy file.
	// A policy with no (0) rules is treated as illegal.
	policyFlag = "--audit-policy-file"
	// specifies the log file path that log backend uses to write audit events. Not specifying this flag disables log backend. - means standard out
	auditLogFlag = "--audit-log-path"
	// defined the maximum number of days to retain old audit log files
	auditLogMaxAge = "--audit-log-maxage"
	// defines the maximum number of audit log files to retain
	auditLogMaxBackup = "--audit-log-maxbackup"
	// defines the maximum size in megabytes of the audit log file before it gets rotated
	auditLogMaxSize = "--audit-log-maxsize"
)

type config struct {
	Namespace         string `env:"NAMESPACE" envDefault:"kube-system"`
	AuditLogMaxAge    int    `env:"AUDIT_LOG_MAX_AGE" envDefault:"1"`
	AuditLogMaxBackup int    `env:"AUDIT_LOG_MAX_BACKUP" envDefault:"2"`
	AuditLogMaxSize   int    `env:"AUDIT_LOG_MAX_SIZE" envDefault:"200"`
}

var (
	cfg config
)

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()
	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// back up apiserver pod yaml
func backupApiServerPodYaml() (string, error) {
	file := fmt.Sprintf("%s%s/%s", nodeRootPath, apiserverManifestDir, apiserverPodConfig)
	backup := fmt.Sprintf("%s%s/../%s.%s", nodeRootPath, apiserverManifestDir, apiserverPodConfig, time.Now().Format(time.RFC3339))

	// back up manifest file
	_, err := copy(file, backup)
	if err != nil {
		log.Printf("failed to copy %s to %s, err: %v\n", file, backup, err)
		// exit error
		return "", err
	}
	return backup, nil
}

// roll back apiserver pod yaml config
func rollBackApiServerPodYaml(backup string) error {
	file := fmt.Sprintf("%s%s/%s", nodeRootPath, apiserverManifestDir, apiserverPodConfig)

	// back up manifest file
	_, err := copy(backup, file)
	if err != nil {
		log.Printf("failed to copy %s to %s, err: %v\n", file, backup, err)
		// exit error
		return err
	}
	return nil
}

// add audit config
func configApiServerAudit() error {
	file := fmt.Sprintf("%s%s/%s", nodeRootPath, apiserverManifestDir, apiserverPodConfig)
	// 读取文件
	b, err := os.ReadFile(file)
	if err != nil {
		log.Print(err)
		return err
	}

	var pod v1.Pod
	// 转换成Struct
	err = yaml.Unmarshal(b, &pod)
	if err != nil {
		log.Printf("%v\n", err.Error())
		return err
	}

	// config command line
	for idx, c := range pod.Spec.Containers {
		container := &pod.Spec.Containers[idx]
		if c.Name == "kube-apiserver" {
			for _, cmd := range container.Command {
				// just skip config when apiserver already has audit flag
				if strings.Contains(cmd, policyFlag) || strings.Contains(cmd, auditLogFlag) {
					return nil
				}
			}
			container.Command = append(container.Command,
				fmt.Sprintf("%s=%s/%s", policyFlag, auditPolicyMountPath, auditPolicyFileName),
				fmt.Sprintf("%s=%s/%s", auditLogFlag, auditLogMountPath, auditLogFileName),
				fmt.Sprintf("%s=%d", auditLogMaxAge, cfg.AuditLogMaxAge),
				fmt.Sprintf("%s=%d", auditLogMaxBackup, cfg.AuditLogMaxBackup),
				fmt.Sprintf("%s=%d", auditLogMaxSize, cfg.AuditLogMaxSize),
			)
		}
	}

	// config audit volume
	policyVolumeName := "audit-policy"
	auditVolumeName := "audit-logs"

	policyType := v1.HostPathFile
	auditLogType := v1.HostPathDirectoryOrCreate

	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: policyVolumeName,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: fmt.Sprintf("%s/%s", auditPolicyHostPath, auditPolicyFileName),
				Type: &policyType,
			},
		},
	},
		v1.Volume{
			Name: auditVolumeName,
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: auditLogHostPath,
					Type: &auditLogType,
				},
			},
		})

	// config volume mount
	for idx, c := range pod.Spec.Containers {
		container := &pod.Spec.Containers[idx]
		if c.Name == "kube-apiserver" {
			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
				Name:      policyVolumeName,
				MountPath: fmt.Sprintf("%s/%s", auditPolicyMountPath, auditPolicyFileName),
				ReadOnly:  true,
			},
				v1.VolumeMount{
					Name:      auditVolumeName,
					MountPath: auditLogMountPath,
				},
			)
		}
	}

	// save to manifest config
	b, err = yaml.Marshal(pod)
	if err != nil {
		log.Printf("failed to marshal pod to yaml bytes , err: %v\n", err)
		// exit error
		return err
	}

	// write manifest file
	err = os.WriteFile(file, b, 0644)
	if err != nil {
		log.Printf("failed to write pod to file , err: %v\n", err)
		// exit error
		return err
	}

	log.Printf("write pod to file %s successfully", file)
	return nil
}

// save policy file to local host
func savePolicyFile() error {
	dir := fmt.Sprintf("%s%s", nodeRootPath, auditPolicyHostPath)

	// make all dir
	err := os.MkdirAll(dir, 0644)
	if err != nil {
		log.Printf("failed to mkdir %s, err: %v\n", dir, err)
		return err
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("failed to fetch kubeconfig in cluster mode, err: %s", err.Error())
		return err
	}

	// fetch configmap
	policyCM, err := clientset.CoreV1().ConfigMaps(cfg.Namespace).Get(context.Background(), auditPolicyConfigMap, metav1.GetOptions{})
	if err != nil {
		log.Printf("failed to fetch audit plicy config map[%s/%s],  err: %s", cfg.Namespace, auditPolicyConfigMap, err.Error())
		return err
	}
	policyContent, ok := policyCM.Data[auditPolicyFileName]
	if !ok || len(policyContent) == 0 {
		log.Printf("can not find audit plicy data from config map[%s/%s],  err: %s", cfg.Namespace, auditPolicyConfigMap, err.Error())
		return err
	}

	// save policy file
	file := fmt.Sprintf("%s/%s", dir, auditPolicyFileName)
	err = os.WriteFile(file, []byte(policyContent), 0644)
	if err != nil {
		log.Printf("failed to write audit policy to file , err: %v\n", err)
		// exit error
		return err
	}

	return nil
}

// touch audit.log and chmod 0644
func createOrPermFile(filepath string) error {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		// create file
		f, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer f.Close()

		// change file permission to 644
		err = os.Chmod(filepath, 0644)
		if err != nil {
			return err
		}
	} else if err == nil {
		// check file permission
		fileInfo, err := os.Stat(filepath)
		if err != nil {
			return err
		}
		permission := fileInfo.Mode().Perm()
		if permission%10 < 4 {
			// change file permission to 644
			err = os.Chmod(filepath, 0644)
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}
	return nil
}

func main() {
	var err error

	if err = env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Printf("cfg is%+v\n", cfg)

	// create audit.log and chmod
	auditFilePath := fmt.Sprintf("%s%s/%s", nodeRootPath, auditLogHostPath, auditLogFileName)
	log.Printf("auditFilePath is %s\n", auditFilePath)
	createOrPermFile(auditFilePath)

	// save policy file to local host
	err = savePolicyFile()
	if err != nil {
		os.Exit(1)
	}

	// back up apiserver original pod config
	var backup string
	backup, err = backupApiServerPodYaml()
	if err != nil {
		os.Exit(1)
	}

	// inject apiserver audit config
	err = configApiServerAudit()
	if err != nil {
		// roll back
		err := rollBackApiServerPodYaml(backup)
		if err != nil {
			os.Exit(1)
		}
	}

	log.Printf("waiting...\n")
	// avoid to exit without any load
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT)
	<-stopChan
}
