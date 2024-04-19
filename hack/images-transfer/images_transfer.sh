#!/bin/bash

# 检查参数数量
if [ "$#" -lt 1 ]; then
    echo "Usage: $0 <target_registry_namespace> [dryrun]"
    exit 1
fi

# 目标镜像的 registry 地址加 namespace
target_registry_namespace=$1
echo "target registry namespace is: $target_registry_namespace"

# 是否为 local 模式
local_mode=false
if [ "$#" -eq 2 ] && [ "$2" == "local" ]; then
    local_mode=true
    echo "Running in local mode. Images will be pulled and tagged locally only."
elif [ "$#" -eq 2 ] && [ "$2" == "dryrun" ]; then
    echo "Running in dryrun mode. Images will not be pulled and pushed."
fi

# 清空输出文件
> /tmp/_out.txt

# 读取 image.txt 文件并进行处理
while IFS= read -r line || [[ -n "$line" ]]; do
    # 忽略注释行和空行
    if [[ $line == \#* ]] || [[ -z $line ]]; then
        continue
    fi

    # 提取镜像名和标签
    image_name=$(echo "$line" | cut -d ':' -f 1)
    image_tag=$(echo "$line" | cut -d ':' -f 2)

    # 如果镜像名中不含有 registry 地址，则默认为 Docker Hub 镜像
    if [[ ! $image_name =~ .*/.* ]]; then
        image_name="registry-1.docker.io/$image_name"
    fi

    # 构建目标镜像名
    target_image="$target_registry_namespace/$(basename $image_name):$image_tag"

    # 输出目标镜像名
    echo "target image is: $target_image"
    # 输出目标镜像名到文件
    echo "$target_image" >> /tmp/_out.txt

    # 如果是 local 模式，则仅拉取源镜像并打标签，不推送到远程仓库
    if $local_mode; then
        # 拉取源镜像
        docker pull "$image_name:$image_tag"
        # 重新打标签
        docker tag "$image_name:$image_tag" "$target_image"
    elif ! $dryrun_mode; then
        # 拉取源镜像
        docker pull "$image_name:$image_tag"
        # 重新打标签
        docker tag "$image_name:$image_tag" "$target_image"
        # 推送目标镜像
        docker push "$target_image"
    fi

done < "images.txt"