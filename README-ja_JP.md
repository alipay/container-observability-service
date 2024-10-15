<p align="center">
<img src="./statics/lunettes-logo.png" width="30%"/>
</p>

# Lunettes - コンテナライフサイクル可観測サービス

<strong><p align="center">スタックを観察し、アプリを活性化する</p></strong>

<p align="center">
<img src="https://img.shields.io/badge/license-Apache 2.0-blue.svg" alt="Apache-2.0 License">
<img src="https://img.shields.io/badge/PRs-Welcome-brightgreen" alt="PRs welcome!" />
</p>


[English](./README.md)

## 🌾 はじめに

Kubernetesはコンテナ・アズ・ア・サービスプラットフォームの構築に広く使用されていますが、コンテナのデリバリープロセスを駆動するために多くの自律的なコンポーネントが協力して動作するため、開発者やSREにとって大きな複雑さを生み出すことがあります。

Lunettesの包括的な可観測サービスは、apiserverリクエストやイベントなどのさまざまな可観測信号を活用して、コンテナライフサイクルのSLI/SLO、診断、およびトレースサービスを作成し、開発者やSREがKubernetes上のサービスをデジタル化された方法で監視および管理できるようにします。

ユーザーフレンドリーなトラブルシューティングとパフォーマンス最適化のアプローチを提供することで、LunettesのソリューションはKubernetes上のサービスの全体的な品質を向上させるのに役立ちます。

## 🔥 主な機能
### リソースデリバリーSLI/SLO:
Lunettesは、インフラストラクチャがコンテナ（Kubernetes上のPod）をデリバリーしようとするのにかかる時間を計算し、このメトリックをコンテナデリバリーSLIとして定義します。このメトリックに基づいて、Lunettesはスケジューリング、イメージプル、IP割り当て、コンテナの起動など、さまざまなコンテナライフサイクルステージに関連する時間コストを認識し、インフラストラクチャの総消費時間を計算できるようにします。一方、コンテナデリバリーSLOは、コンテナの仕様に基づいて定義されます。

LunettesのコンテナデリバリーSLI/SLOの定義により、サービスオーナーはプラットフォームのリソースデリバリープロセスの品質をデジタル化された方法で評価および改善することができます。

![ContainerDeliverySli/Slo](./statics/deliveryslo.png)
### コンテナライフサイクル診断サービス
Lunettesは、コンテナライフサイクル全体の可観測信号を分析し、過剰なリソース消費エラー、構成エラーなどの一般的な問題をカバーするエラーコードを割り当てることで、問題の根本原因を特定します。

![ContainerDeliverySli/Slo](./statics/deliverydiagnose.png)

### コンテナライフサイクルトレースサービス
Lunettesは、各コンテナライフサイクルステージの開始と終了を認識することにより、OpenTelemetry標準に従ったトレース構造を構築します。

![ContainerDeliverySli/Slo](./statics/deliverytracing.png)

## 🎬 始める

### クイックスタート

[こちらのガイド](./docs/quick_start.md)を参照して、[kind](https://kind.sigs.k8s.io/)を使用して迅速に開始します。

### デプロイ
ステップ1: Kubeadm/Kindを使用してKubernetesクラスターをブートストラップします。

- [kubeadmを使用してクラスターを作成する](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helmのインストール](https://helm.sh/docs/intro/install/)

次の方法では、サービスをNodePortを介して公開します。現在の操作環境がKubernetesのnodeIPにアクセスできることを確認してください。

ステップ2: Helmを使用してLunettesをインストールします

注意: Helm v3.8.0以降、OCIサポートがデフォルトで有効になり、実験的から一般提供に昇格しました。そのため、Helm v3.8.0以上を選択することをお勧めします。

```bash
# NodePortを使用する
helm install lunettes oci://registry-1.docker.io/lunettes/lunettes-chart --version [version] \
  # enableAuditApiserverをtrueに設定すると、apiserverの監査が有効になります。
  # このプロセスはapiserverを再起動しますので注意してください。
  --set enableAuditApiserver=true \
  --set grafanaType=NodePort \
  --set jaegerType=NodePort 
```
利用可能な[バージョン](https://hub.docker.com/r/lunettes/lunettes-chart/tags)を参照してください

ステップ3: Lunettesダッシュボードサービスのエンドポイントを見つける
```bash
export LUNETTES_IP=node_ip
export GRAFANA_NODEPORT=$(kubectl -n lunettes get svc grafana -o jsonpath='{.spec.ports[0].nodePort}')
export JAEGER_NODEPORT=$(kubectl -n lunettes get svc jaeger-collector -o jsonpath='{.spec.ports[0].nodePort}')
```

ブラウザで[http://[LUNETTES_IP]:[GRAFANA_NODEPORT]](http://[LUNETTES_IP]:[GRAFANA_NODEPORT])を開き、debugpodまたはdebugsloエンドポイントにアクセスします。デフォルトのユーザー名とパスワードは`admin:admin`です。

ブラウザで[http://[LUNETTES_IP]:[JAEGER_NODEPORT]/search?](http://[LUNETTES_IP]:[JAEGER_NODEPORT]/search?)を開き、トレースエンドポイントにアクセスします。

## 🛠 設定
Lunettesは高度に構成可能です。以下に、リソースデリバリーSLOおよびコンテナライフサイクルトレースをさまざまなシナリオに適応させるための簡単な設定例を示します。
### リソースデリバリーSLOの設定
```json
{
    "UserOnlineConfigMap":{
        "test-ns-one":"1m30s",
        "test-ns-two":"6m"
    },
    "IgnoredNamespaceForAudit":[
        "app-ns"
    ],
    "IgnoreDeleteReasonNamespace":[
        "test-ns-three",
        "test-ns-four"
    ]
}
```
### コンテナライフサイクルトレースの設定
```json
[
  {
    "ObjectRef":{
      "Resource":"pods",
      "Name":"PodSpans",
      "APIVersion":"v1"
    },
    "ActionType":"PodCreate",
    "LifeFlag":{
      "Mode":"start-finish",
      "StartEvent":[
        {
          "Type":"operation",
          "Operation":"pod:create:success"
        }
      ],
      "FinishEvent":[
        {
          "Type":"operation",
          "Operation":"condition:Ready:true"
        }
      ]
    },
    "ExtraProperties":{
      "bizName":{
        "Name":"",
        "ValueRex":"metadata#labels#meta.k8s.com/biz-name",
        "NeedMetric":true
      }
    },
    "Spans":[
      {
        "Name":"default_schedule_span",
        "Type":"default_schedule_span",
        "SpanOwner":"k8s",
        "Mode":"start-finish",
        "StartEvent":[
          {
            "Type":"operation",
            "Operation":"schedule:default-scheduler:entry"
          }
        ],
        "EndEvent":[
          {
            "Type":"operation",
            "Operation":"schedule:binding:success"
          }
        ]
      }
    ]
  }
]
```


## 📑 ドキュメント
詳細については、[docs](/docs)をご覧ください

## 💡 コミュニティ
Lunettesに関連する質問がある場合は、以下の方法でお問い合わせください：
- Slack
- [DingTalk](https://qr.dingtalk.com/action/joingroup?code=v1,k1,11giuLFUSQIWJ1Otmzn2egYQnu9p+sNhe1yktypjpz0=&_dt_no_comment=1&origin=11)
- [GitHub Issue](https://github.com/alipay/container-observability-service/issues)
