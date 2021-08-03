# sts-annotator <!-- omit in toc -->

- [How it works](#how-it-works)
  - [Limitations](#limitations)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Quick Install](#quick-install)
  - [Custom Install](#custom-install)
- [Configurations](#configurations)
  - [sts-annotator configurations](#sts-annotator-configurations)
  - [sts-annotator deployment configurations](#sts-annotator-deployment-configurations)
  - [MutatingWebhookConfiguration Configurations](#mutatingwebhookconfiguration-configurations)
  - [TLS Certificate Configurations](#tls-certificate-configurations)
- [Tips](#tips)
  - [Generating TLS certificate](#generating-tls-certificate)

sts-annotator is Kubernetes admission controllers, which is an API providing custom admission review in your cluster. The
goal of this controller is to allow pods to get `POD_INDEX` and `POD_REPLICAS` from Environment
Variables. `sts-annotator` achieve it by adding annotations to any pods managed by Statefulset at the time of `CREATE`
and `UPDATE`. The `POD_INDEX` is the pod's ordinal index suffixed to the Statefulset Pod Name.

If you want to understand basics,
read [A Guide to Kubernetes Admission Controllers](https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers/)
. If you want to understand required specifications over the custom APIs,
read [Dynamic Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
.

# How it works

`sts-annotator` add annotations to any pods managed by Statefulset at the time of CREATE or UPDATE. It looks like below.

**A Statefulset POD (api-somethings-v1-0) is successfully annotated**

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: api-somethings-v1-sts-0
  namespace: default
  annotations:
    sts-annotator/pod-index: "0"
    sts-annotator/pod-replicas: "2"
```

You can get the annotations at your Environment variable, accessing to the annotations by `fieldRef`.

**How to get the pod-index and pod-replicas from Env Vars**

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: "api-somethings-v1-sts"
spec:
  template:
    spec:
      containers:
        - env:
            - name: POD_INDEX
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['sts-annotator/pod-index']
            - name: POD_REPLICAS
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['sts-annotator/pod-replicas']
```

You can see the Environment variables correctly loaded when ssh into the pod.

**A Statefulset POD (api-somethings-v1-0) is successfully load env vars**

```sh
# POD: api-somethings-v1-sts-0, Statefulset Replicas: 3
$ env | grep POD
POD_REPLICAS=2
POD_INDEX=0
```

## Limitations

`sts-annotator` can add annotations to pods when they are initially created and never update it. So that it won't update the `metadata.annotations['sts-annotator/pod-replicas']` of existing pods when you change the number of replicas in Statefulset unless you recreate all pods. If this is critical for your application, please raise an issue. We are thinking to add an option of auto remediation, which will re-create all pods under the updated Statefulset.

# Installation

## Prerequisites

- k8s cluster is running

## Quick Install

We used to use a set of static yaml files to deploy `sts-annotator` but Helm is supported now and recommend you to use
Helm to install `sts-annotator` now.

If you want to quickly test `sts-annotator`, please follow the steps below. The quick install uses the pre-generated certificate and requires sts-annotator deployed in `sts` namespace.

```sh
# clone source code and move into the directory
$ git clone https://github.com/mk811/sts-annotator.git
$ cd sts-annotator

# the namespace name must be sts. Other namespace name is not supported by the TLS certificate
$ kubectl create ns sts
$ helm repo add mk811 https://raw.githubusercontent.com/mk811/sts-annotator/master/chart
$ helm install test mk811/sts-annotator  -f test/values.yaml -n sts
```

## Custom Install

You can install `sts-annotator` with your own configurations.

To deploy `sts-annotator` to your namespaces, you need to generate a server certificate for sts-annotator by yourself. This is simply because `MutatingWebhookConfiguration` sends TLS encrypted request to `sts-annotator`. Please see [Generating TLS certificate](#generating-tls-certificate) to find the way to generate TLS certificate.

Once you get cert, key, certificate authority, please create a `values.yaml` file in your local machine and set base64 encoded cert, key and CA in the file. Also, you can add more configurations as you like. Please see the [Configurations](#configurations) section for available parameters.

```yaml
# values.yaml
tls:
  cert: "LS0tLS1..."
  key: "LS0tL..."
  caCert: "LS0tL..."
```

Now, let's move onto the installation.

```sh
# Create a namespace for the sts-annotator.
$ kubectl create namespace <your-namespace>

# In this step, helm uses the values.yaml created step above.
$ helm repo add mk811 https://raw.githubusercontent.com/mk811/sts-annotator/master/chart
$ helm install <name e.g. `test`> mk811/sts-annotator -f values.yaml -n <your-namespace>

NAME: test
LAST DEPLOYED: Mon Aug  2 17:49:58 2021
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

# Configurations

Default values for all configurations are defined in [values.yaml](chart/sts-annotator/values.yaml).

## sts-annotator configurations

`config` is a section for configuring sts-annotator behavior.

| field            | description                                                         |
| ---------------- | ------------------------------------------------------------------- |
| server.mode      | if you set 'debug', the api will run with 'debug' mode.             |
| server.log.level | log level, either of 'error' , 'info' or 'debug',                   |
| server.port      | port where the `sts-annotator` listens to                           |
| server.tls.cert  | path to the cert file for `sts-annotator.<namespace>.svc.cluster`   |
| server.tls.key   | path to the private key for `sts-annotator.<namespace>.svc.cluster` |
| k8s.url          | combination of url and port where your master api listens to        |
| k8s.token        | path to the k8s token file, used to call master api                 |
| k8s.tls.caCert   | path to the ca file, used to call master api                        |

## sts-annotator deployment configurations

`app` is a section for configuring sts-annotator deployment configurations

| field           | description                                                   |
| --------------- | ------------------------------------------------------------- |
| replicas        | number of replicas of sts-annotator                           |
| labels          | optional labels for sts-annotator deployment and pod          |
| image           | sts-annotator image                                           |
| tag             | sts-annotator image version                                   |
| env             | default set of environment variables                          |
| imagePullPolicy | k8s imagePullPolicy for sts-annotator deployment              |
| resources       | k8s resources for sts-annotator deployment                    |
| affinity        | k8s affinity configuration for sts-annotator deployment       |
| livenessProbe   | k8s livenessProbe configuration for sts-annotator deployment  |
| readinessProbe  | k8s readinessProbe configuration for sts-annotator deployment |
| volumes         | k8s volume configuration for sts-annotator deployment         |
| volumeMounts    | k8s volumeMounts configuration for sts-annotator deployment   |

## MutatingWebhookConfiguration Configurations

`mutationWebhook` is a section for configuring `MutatingWebhookConfiguration` resource.

| field         | description                                                                                                                                     |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| apiVersion    | `apiVersion` of `MutatingWebhookConfiguration`. `v1` or `v1beta`. `v1` is recommended as it is latest version.                                  |
| failurePolicy | failurePolicy defines how unrecognized errors and timeout errors from the admission webhook are handled. Allowed values are `Ignore` or `Fail`. |
| sideEffects   | The .webhooks[].sideEffects field should be set to None if a webhook doesn't have any side effect.                                              |

## TLS Certificate Configurations

`tls` is a section for configuring TLS certificate used within `sts-annotator` server. Default values for the fields under `TLS` section aren't defined as it should be different on your environment.

| field  | description                                    |
| ------ | ---------------------------------------------- |
| cert   | BASE64 Encoded Certificate                     |
| key    | BASE64 Encoded Private Key for the Certificate |
| caCert | BASE64 Encoded CA Certificate                  |

# Tips

## Generating TLS certificate

If you have k8s admin permission (or appropriate permissions), you can generate valid server certificate for `sts-annotator` in k8s cluster. Please refer to [k8s official docs](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/) for details. The following steps are the sample commands used to generate certificate [test/test.crt](test/test.crt) and [test.key](test/test.key).

Note, we verified the following commands on MacOS only. `base64` or other MacOS native commands may be vary if you are using Windows or Linux systems.

1. Create a Certificate Signing Request

   Please update `hosts` for your namespace where the sts-annotator is deployed. The sample below expects `sts-annotator` is deployed in `sts` namespace. If you deploy it to `default` namespace, the `hosts` should be: `["sts-annotator.default.svc.cluster.local", "sts-annotator.default.svc", "sts-annotator"]`.

   ```sh
   cat <<EOF | cfssl genkey - | cfssljson -bare server
   {
     "hosts": [
       "sts-annotator.sts.svc.cluster.local",
       "sts-annotator.sts.svc",
       "sts-annotator"
     ],
     "CN": "system:node:sts-annotator.sts.pod.cluster.local",
     "key": {
       "algo": "ecdsa",
       "size": 256
     },
     "names": [
       {
         "O": "system:nodes"
       }
     ]
   }
   EOF
   ```

2. Create a Certificate Signing Request object to send to the Kubernetes API

   ```sh
   cat <<EOF | kubectl apply -f -
   apiVersion: certificates.k8s.io/v1
   kind: CertificateSigningRequest
   metadata:
     name: sts-annotator
   spec:
     request: $(cat server.csr | base64 | tr -d '\n')
     signerName: kubernetes.io/kubelet-serving
     usages:
     - digital signature
     - key encipherment
     - server auth
   EOF
   ```

   This command generates two files; it generates `server.csr` containing the PEM encoded pkcs#10 certification request, and `server-key.pem` containing the PEM encoded key to the certificate that is still to be created.

   Let's base64 encode the key as we will use it in installation.

   ```sh
   $ cat server-key.pem | base64 > server-key.base64.pem
   ```


   Also, now you should be able to see the certificate signing request.

   ```sh
   kubectl get csr sts-annotator
   ```

3. Approve CSR

   If you have admin permission, you can approve the CSR by command below.

   ```sh
   kubectl certificate approve sts-annotator
   ```

4. Get server certificate (base64 encoded)

   ```sh
   kubectl get csr sts-annotator -o jsonpath='{.status.certificate}' > server.base64.crt
   ```

   Now, you can find the `server-key.pem` in your local and you can download certificate.

5. Get Certificate Authority signed the Certificate. (If you don't know how to get the Certificate Authority)

   Here are some tips to get Kubernetes Certificate Authority.

   1. get through kubectl command  
      There is a chance to get Certificate Authority by the command below. However, you might not be able to get it. (At
      least, I couldn't get CA for some environments)

   ```sh
   kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' > ca.base64.crt
   ```

   2. retrieve from pod  
      Pick up any pod you prefer and run the command below. You can copy the Kubernetes Master API Certificate Authority
      from POD.

   ```sh
   POD_NAME="<your/pod/name>"
   NAMESPACE="<your/namespace>"
   kubectl exec sts-annotator-869bb7f4b-9cdc6  -- cat "/run/secrets/kubernetes.io/serviceaccount/ca.crt" | base64 > ca.base64.crt
   ```

   It is also okay to generate new certificates by
   following [here](https://kubernetes.io/docs/concepts/cluster-administration/certificates/) or you can go with the steps above. In either case, please discuss the certificate lifecycle management with your system admin for production use.

6. Done!
   Finally, you have server certificate, key and CA in your local.

   ```sh
   $ ls | grep base64
   ca.base64.crt
   server-key.base64.pem
   server.base64.crt
   ```

   
