
git clone https://github.com/istio/client-go.git
cd client-go/cmd/example
go get .
go run .
export KUBECONFIG='/root/.kube/config' && export NAMESPACE='default' && go run client.go


How to Run
go get .
export KUBECONFIG='/root/.kube/config' && export NAMESPACE='istio-test' && go run .
swag init ./main.go

go RESTful API totorial
https://go.dev/doc/tutorial/web-service-gin
go compile tutorial
https://qiita.com/uchiko/items/64fb3020dd64cf211d4e

Istio Client-go -- VirtualServices Traffic Shifting Weights
https://blog.csdn.net/qq_29778131/article/details/108198446?spm=1001.2101.3001.6650.1&utm_medium=distribute.pc_relevant.none-task-blog-2%7Edefault%7EBlogCommendFromBaidu%7ERate-1.pc_relevant_default&depth_1-utm_source=distribute.pc_relevant.none-task-blog-2%7Edefault%7EBlogCommendFromBaidu%7ERate-1.pc_relevant_default&utm_relevant_index=2

https://istio.io/latest/docs/tasks/traffic-management/traffic-shifting/
测试配置 patch vs
https://blog.csdn.net/baobaoxiannv/article/details/110732147
https://www.codercto.com/a/31924.html
canary description
https://istio.io/latest/blog/2017/0.1-canary/


curl http://localhost:3011/api/icg/weightConfig \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"id": "4","title": "The Modern Sound of Betty Carter","artist": "Betty Carter","price": 49.99}'


curl http://localhost:3011/api/icg/weightConfig \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{ "ns": "istio-test", "svcname": "customers", "versions": ["v1","v2"], "weights": [30, 70]}'

kubectl -n istio-test get vs reviews   -o yaml

swaggerUI setup
https://qiita.com/takehanKosuke/items/bbeb7581330910e72bb2
Goの実行時にimportエラーでハマった話。
https://qiita.com/Qii_Takuma/items/bf2aefe066ea616c6c72

swag init command not found
https://blog.csdn.net/weixin_43262264/article/details/107339026
export PATH="/root/go/bin:$PATH"
swag init