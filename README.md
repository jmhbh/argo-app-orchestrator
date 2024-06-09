# argo-app-orchestrator

This project is a demo of how to programatically use the ArgoCD ApplicationSet to manage deployments for multiple users. As well as integrate the argo managed services into a tailnet network using the tailnet k8s operator.

## Setup
1. Install `microk8s` following the instructions [here](https://ubuntu.com/tutorials/install-a-local-kubernetes-with-microk8s#2-deploying-microk8s) or
if you're running on an M1 Mac, install with `brew install ubuntu/microk8s/microk8s`.
2. Setup `microk8s` with the following commands. Note: if you have issues with multipass you may need to run `multipass authenticate`:
```shell
microk8s install
microk8s enable community
microk8s enable argocd
microk8s enable helm
```
3. Update your kubeconfig with the following command:
```shell
microk8s kubectl config view --raw > $HOME/.kube/config
```
4. Ensure you have your `tailscale` OAuthClientID and OAuthClientSecret created and set as env vars. You can get these by following the instructions [here](https://tailscale.com/kb/1215/oauth-clients). Note that for this demo the OAuth Client has full read and write scopes. Make sure to create the appropriate ACL tags in your tailnet policy file as documented in the prerequisites section [here](https://tailscale.com/kb/1236/kubernetes-operator). Cone you have completed the previous steps now install the tailscale operator by running the following command:
```shell
helm upgrade \
  --install \
  tailscale-operator \
  tailscale/tailscale-operator \
  --namespace=tailscale \
  --create-namespace \
  --set-string oauth.clientId=$TAILNET_OAUTH_CLIENT_ID \
  --set-string oauth.clientSecret=$TAILNET_OAUTH_CLIENT_SECRET \
  --wait
```
5. Deploy the `argo-app-orchestrator` application by running the following commands in the project root directory. The multiarch docker image for this application is available on my public docker repo [here](https://hub.docker.com/r/jmhbh/public/tags):
```shell
microk8s kubectl apply -f ./setup/setup.yaml
```

## Usage

1. Port forward the webserver on port 9000 after the pod is running:
```shell
microk8s kubectl port-forward service/app-orchestrator -n app-orchestrator 9000:9000
```
2. Curl the webserver with the following request:
```shell
curl -X POST \
  http://localhost:9000/api/v1/create \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "test",
    "email": "test@test.com"
}'
```
3. You should see the following response:
```
Modified Argo ApplicationSet to manage deployment for additional user: test
```
4. Curl the webserver with another request for an additional user:
```shell
curl -X POST \
  http://localhost:9000/api/v1/create \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "test2",
    "email": "test2@test.com"
}'
```
5. You should see the following response:
```
Modified Argo ApplicationSet to manage deployment for additional user: test2
```
6. You can check that the ApplicationSet has been created and should look like the following. Note that the list generator user is populated with the user metadata provided in the post request. The user will have their own namespace created in the `{{user}}-app` format. And the Application deploys the resources (A super mario deployment, and a tailscale server) in my git-ops repo [here](https://github.com/jmhbh/git-ops).
```shell
microk8s kubectl get appsets -A -oyaml
```
```yaml
apiVersion: v1
items:
  - apiVersion: argoproj.io/v1alpha1
    kind: ApplicationSet
    metadata:
      creationTimestamp: "2024-06-09T00:49:40Z"
      generation: 2
      name: supermario-app-set
      namespace: argocd
      resourceVersion: "4164"
      uid: dfe555f8-ddcf-404f-8b41-cbd5df576907
    spec:
      generators:
        - list:
            elements:
              - user: test
              - user: test2
      template:
        metadata:
          name: '{{user}}-app'
        spec:
          destination:
            namespace: '{{user}}-app'
            server: https://kubernetes.default.svc
          project: default
          source:
            path: example
            repoURL: https://github.com/jmhbh/git-ops.git
            targetRevision: HEAD
          syncPolicy:
            automated:
              prune: false
              selfHeal: true
            syncOptions:
              - ApplyOutOfSyncOnly=true
              - CreateNamespace=true
```
5. Check the Application created by the ApplicationSet with the following command if you wish to do so:
```shell
microk8s kubectl get applications -A -oyaml
```
6. Navigate to the namespace created by argo for the user and exec into the tailscale pod to enable ssh.
```shell
❯ microk8s kubectl get pods -n test-app
NAME                         READY   STATUS    RESTARTS      AGE
supermario-cdf945587-k2wq2   1/1     Running   0             4m44s
tailscale-7f57d8c5d8-hcqnj   1/1     Running   3 (79s ago)   4m44s

❯ microk8s kubectl exec -it -n test-app tailscale-7f57d8c5d8-hcqnj -- /bin/sh
/ # tailscale login

To authenticate, visit:

	https://login.tailscale.com/a/142da15101efa9
Success.
/ # tailscale set --ssh
```
7. The tailscale pod is automatically added to your tailnet by the tailscale k8s operator and you can now ssh into the pod as well. If you navigate to your tailscale admin console [here](https://login.tailscale.com/admin/machines) you should see the machine name for your pod there.
8. Try sshing into the pod you can use the ssh console in the tailscale admin console to do so.
9. Lastly don't forget to have some fun. Port forward the supermario service and play some super mario at http://localhost:8600/!
```shell
microk8s kubectl port-forward service/supermario -n test-app 8600:8080
```
