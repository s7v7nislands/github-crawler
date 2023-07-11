# github-crawler

github-crawler get user info in github

## create github OAuth app

Follow github doc: [Creating an OAuth app](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app)
and get client_id and client_secret of the app. You can put them into `.env`

```bash
ClIENT_ID="xxxx"
CLIENT_SECRET="xxxx"
```

## helm install

dev in local machine, need install:

* minikube or any k8s cluster
* kubectl
* helm

you can follow the steps:

```bash
# export ClIENT_ID and CLIENT_SECRET
source .env

cd helm
helm install --set clientID=$ClIENT_ID --set clientSecret=$CLIENT_SECRET crawler ./github-crawler

kubectl --namespace default port-forward svc/crawler-grafana 3000:80
# grafana login user/password: admin/prom-operator
```
