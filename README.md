# k8s-extended-apiserver
Kubernetes Extended APIServer using net/http library
Resource:
1. https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
2. https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/
3. https://kubernetes.io/docs/tasks/extend-kubernetes/configure-aggregation-layer/
4. https://medium.com/@vanSadhu/kubernetes-api-aggregation-setup-nuts-bolts-733fef22a504
5. https://github.com/tamalsaha/DIY-k8s-extended-apiserver/tree/master [***]
6. https://youtu.be/pTIwy6fpxwY [***]


Execution Process List:
1. RUN Extended-Server
2. RUN Api Server
3. RUN API Aggregation call

```console
# Terminal 1
$ go run api-server/main.go
2023/07/20 15:36:37 listening on 127.0.0.1:8443

# Terminal 2
export APISERVER_ADDR=127.0.0.1:8443

$ curl -k https://127.0.0.1:8443/api-server/pods
Resource: pods
```

```console
# Terminal 1
$ go run extended-apiserver/main.go
2023/07/20 15:45:46 listening on 127.0.0.2:8443

# Terminal 2
export EAS_ADDR=127.0.0.2:8443

$ curl -k https://127.0.0.2:8443/extended-apiserver/mariadb
Resource: mariadb requested by user[-]=system:anonymous
```

```console
# Terminal 1
$ go run api-server/main.go --send-proxy-request=true
forward true
2023/07/20 15:48:42 listening on 127.0.0.1:8443

# Terminal 2
$ go run extended-apiserver/main.go --receive-proxy-request=true
forward true
2023/07/20 15:48:38 listening on 127.0.0.2:8443

# Terminal 3

$ curl -k https://127.0.0.1:8443/api-server/pods
Resource: pods

$ curl -k https://127.0.0.2:8443/extended-apiserver/mariadb
Resource: mariadb requested by user[-]=system:anonymous

$ curl -k https://127.0.0.1:8443/extended-apiserver/mariadb
Resource: tmrbal requested by user[X-Remote-User]=

$ `curl https://127.0.0.1:8443/extended-apiserver/mariadb --cacert certs/api-server-ca.crt --cert certs/api-server-uba.crt --key certs/api-server-uba.key`
Resource: mariadb requested by user[X-Remote-User]=uba

```