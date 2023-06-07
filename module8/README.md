# 作业八

## 作业要求

编写 Kubernetes 部署脚本将 httpserver 部署到 Kubernetes 集群，以下是你可以思考的维度。

------

第一部分

- 优雅启动
- 优雅终止
- 资源需求和 QoS 保证
- 探活
- 日常运维需求，日志等级
- 配置和代码分离

------

第二部分

- Service
- Ingress
- 如何确保整个应用的高可用
- 如何通过证书保证 httpServer 的通讯安全

## 作业步骤

1	创建deployment

```sh
root@yon:~/cncamp/module8# kubectl apply -f deployment.yaml
```

2	创建创建configmap

```sh
root@yon:~/cncamp/module8# kubectl apply -f configmap.yaml
```

3	创建创建service 

```sh
root@yon:~/cncamp/module8# kubectl apply -f service.yaml
```

4	查看pod

```sh
root@yon:~/cncamp/module8# kubectl get pod
NAME                          READY   STATUS    RESTARTS   AGE
httpserver-584b4bcdb6-cqw24   1/1     Running   0          25h
httpserver-584b4bcdb6-km2xs   1/1     Running   0          25h
httpserver-584b4bcdb6-vsxpw   1/1     Running   0          25h
```

5	查看endpoints

```sh
root@yon:~/cncamp/module8# kubectl get ep
NAME         ENDPOINTS                                       AGE
httpsvc      10.0.0.163:8080,10.0.0.52:8080,10.0.0.92:8080   25h
```

6	查看service

```sh
root@yon:~/cncamp/module8# kubectl get svc -n ingress-nginx
NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   	PORT(S)     AGE
ingress-nginx-controller             NodePort    10.98.138.155   <none>        80:31891/TCP  8h
ingress-nginx-controller-admission   ClusterIP   10.111.51.160   <none>        443/TCP       8h
```

7    NodePort访问

```sh
root@yon:~/cncamp/module8# curl http://10.98.138.155:31891/healthz
working
```

8	安装ingress

```sh
root@yon:~/cncamp/module8# kubectl apply -f ingress-deploy.yaml
```

9	查看ingress安装情况 

```sh
root@yon:~/cncamp/module8# kubectl get pod -n ingress-nginx
NAME                                        READY   STATUS      RESTARTS   AGE
ingress-nginx-admission-create-njq4l        0/1     Completed   0          8h
ingress-nginx-admission-patch-98mrv         0/1     Completed   0          8h
ingress-nginx-controller-667747967b-9kb78   1/1     Running     0          8h
```

10	创建ingress

```sh
root@yon:~/cncamp/module8# kubectl apply -f ingress.yaml
```

11 查看ingress

```sh
root@yon:~/cncamp/module8# kubectl get ingress
Warning: extensions/v1beta1 Ingress is deprecated in v1.14+, unavailable in v1.22+; use networking.k8s.io/v1 Ingress
NAME            CLASS    HOSTS                ADDRESS          PORTS     AGE
https-ingress   <none>   www.yon.com          192.168.126.55   80, 443   24m
```

12 测试ingress(做本地hosts解析)

```sh
root@master:~# curl http://www.yon.com:31891/healthz
working
```

13 

```bash
root@yon:~/cncamp/module8# cat yon-csr.json 
{
    "CN": "yon.com",
    "hosts": [
        "www.yon.com"
    ],
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "CN",
            "L": "Shanghai",
            "ST": "Shanghai"
        }
    ]
}
root@yon:~/cncamp/module8# kubectl create secret tls yon-secret --cert=ca.pem --key=ca-key.pem
```

14 查看secret

``` sh
root@yon:~/cncamp/module8# kubectl get secret
NAME                  TYPE                                  DATA   AGE
default-token-bf488   kubernetes.io/service-account-token   3      31h
yon-secret            kubernetes.io/tls                     2      21m
```

15	创建https-ingress

```  sh
root@yon:~/cncamp/module8# kubectl apply -f https-ingress.yaml
```

16	查看ingress

```  sh
root@yon:~/cncamp/module8# kubectl get ingress
Warning: extensions/v1beta1 Ingress is deprecated in v1.14+, unavailable in v1.22+; use networking.k8s.io/v1 Ingress
NAME            CLASS    HOSTS                ADDRESS          PORTS     AGE
https-ingress   <none>   www.yon.com          192.168.126.55   80, 443   28m
```

17	测试证书是否生效

``` sh
root@yon:~/cncamp/module8# curl  https://www.yon.com/healthz
curl: (60) SSL certificate problem: unable to get local issuer certificate
More details here: https://curl.haxx.se/docs/sslcerts.html

curl failed to verify the legitimacy of the server and therefore could not
establish a secure connection to it. To learn more about this situation and
how to fix it, please visit the web page mentioned above.

# 31173 ingress 的 nodeport 端口
root@yon:~/cncamp/module8# curl -kv https://www.yon.com:31173
*   Trying 192.168.126.55:31173...
* TCP_NODELAY set
* Connected to www.yon.com (192.168.126.55) port 31173 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/certs/ca-certificates.crt
  CApath: /etc/ssl/certs
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
* TLSv1.3 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-AES128-GCM-SHA256
* ALPN, server accepted to use h2
* Server certificate:  # 可以看到使用了自己的证书
*  subject: C=CN; ST=Shanghai; L=Shanghai; CN=www.yon.com
*  start date: Jun  7 01:02:50 2023 GMT
*  expire date: Jun  6 01:02:50 2024 GMT
*  issuer: C=CN; ST=Shanghai; L=Shanghai; CN=www.yon.com
*  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x55de8ec0a300)
> GET / HTTP/2
> Host: www.yon.com:31173
> user-agent: curl/7.68.0
> accept: */*
> 
* Connection state changed (MAX_CONCURRENT_STREAMS == 128)!
< HTTP/2 200 
< server: nginx/1.17.10
< date: Wed, 07 Jun 2023 09:28:37 GMT
< content-length: 0
< accept: */*
< user-agent: curl/7.68.0
< version: v0.0.1
< x-forwarded-for: 10.0.0.235
< x-forwarded-host: www.yon.com:31173
< x-forwarded-port: 443
< x-forwarded-proto: https
< x-real-ip: 10.0.0.235
< x-request-id: e9a0062f82c037658bd0d67a1f775ed1
< x-scheme: https
< strict-transport-security: max-age=15724800; includeSubDomains
< 
* Connection #0 to host www.yon.com left intact
```

