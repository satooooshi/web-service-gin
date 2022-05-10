package main

import (
	"context"
	"log"
	"net/http"
	"os"

	_ "web-service-gin/docs" // ←追記

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"

	// for apply
	//clientgonetworkingv1alpha3 "istio.io/client-go/pkg/applyconfiguration/networking/v1alpha3" // https://pkg.go.dev/istio.io/client-go@v1.13.3/pkg/applyconfiguration/networking/v1alpha3#VirtualServiceApplyConfiguration
	//_ "istio.io/client-go/pkg/applyconfiguration/meta/v1"                              // https://pkg.go.dev/istio.io/client-go@v1.13.3/pkg/applyconfiguration/meta/v1#TypeMetaApplyConfiguration

	// defineLBConfig
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// swagger UI
	// docsのディレクトリを指定

	ginSwagger "github.com/swaggo/gin-swagger"   // ←追記
	"github.com/swaggo/gin-swagger/swaggerFiles" // ←追記
)

// ac "istio.io/client-go/pkg/applyconfiguration"
// istio.io/client-go/pkg/applyconfiguration/utils.go
// /Users/satoshiaikawa/client-go-master/pkg/applyconfiguration/utils.go

type weights struct {
	Ns       string   `json:"ns"`
	Svcname  string   `json:"svcname"`
	Versions []string `json:"versions"`
	Weights  []int32  `json:"weights"`
}

type lb struct {
	Ns      string `json:"ns"`
	Svcname string `json:"svcname"`
	Lb      string `json:"lb"`
}

// getAlbums responds with the list of all albums as JSON.
func getExample(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "Hello Istio Client Go")
}

// getIstioConfig responds with the list of all as JSON.
// @Summary lists istio configurations of intio-gateway, virtual service, and destination rules.
// @Tags Istio Resouce Config
// @Accept  json
// @Produce  json
// @Success 200 {string} string	"ok"
// @Failure 400
// @Failure 404
// @Router /api/icg/getIstioConfig [get]
func getIstioConfig(c *gin.Context) {
	kubeconfig := os.Getenv("KUBECONFIG") // os.GEtenv gets environment variable
	namespace := os.Getenv("NAMESPACE")

	if len(kubeconfig) == 0 || len(namespace) == 0 {
		log.Fatalf("Environment variables KUBECONFIG and NAMESPACE need to be set")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create k8s rest client: %s", err)
	}

	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	// Test VirtualServices
	vsList, err := ic.NetworkingV1alpha3().VirtualServices(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get VirtualService in %s namespace: %s", namespace, err)
	}

	for i := range vsList.Items {
		vs := vsList.Items[i]
		log.Printf("Index: %d VirtualService Hosts: %+v\n", i, vs.Spec.GetHosts())
		log.Printf("%+v\n", vs.Spec.GetHttp()[0].GetRoute()[0].GetWeight())
	}

	// Test DestinationRules
	drList, err := ic.NetworkingV1alpha3().DestinationRules(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get DestinationRule in %s namespace: %s", namespace, err)
	}

	for i := range drList.Items {
		dr := drList.Items[i]
		log.Printf("Index: %d DestinationRule Host: %+v\n", i, dr.Spec.GetHost())
	}

	// Test Gateway
	gwList, err := ic.NetworkingV1alpha3().Gateways(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get Gateway in %s namespace: %s", namespace, err)
	}

	for i := range gwList.Items {
		gw := gwList.Items[i]
		for _, s := range gw.Spec.GetServers() {
			log.Printf("Index: %d Gateway servers: %+v\n", i, s)
		}
	}

	// Test ServiceEntry
	seList, err := ic.NetworkingV1alpha3().ServiceEntries(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get ServiceEntry in %s namespace: %s", namespace, err)
	}

	for i := range seList.Items {
		se := seList.Items[i]
		for _, h := range se.Spec.GetHosts() {
			log.Printf("Index: %d ServiceEntry hosts: %+v\n", i, h)
		}
	}
	//c.IndentedJSON(http.StatusOK, "Get Istio Config")
	c.IndentedJSON(http.StatusOK, "Get Istio Config")
}

// postWeightConfig
// @Summary defines weight policies that apply to traffic intended for a service after routing has occurred.
// @Tags Istio Resouce Config
// @Accept  json
// @Produce  json
// @Param title body string true "title"
// @Param body body string true "body"
// @Success 201
// @Failure 400
// @Router /api/icg/weightConfig [post]
func postWeightConfig(c *gin.Context) {

	var newWeights weights

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newWeights); err != nil {
		return
	}

	log.Printf("posted weights cifig: %+v\n", newWeights)
	log.Printf("namespace: %+v\n", newWeights.Ns)
	log.Printf("service name: %+v\n", newWeights.Svcname)
	log.Printf("versions: %+v\n", newWeights.Versions)
	log.Printf("weights: %+v\n", newWeights.Weights)

	namespace := newWeights.Ns
	kubeconfig := os.Getenv("KUBECONFIG") // os.GEtenv gets environment variable

	if len(kubeconfig) == 0 || len(namespace) == 0 {
		log.Fatalf("Environment variables KUBECONFIG and NAMESPACE need to be set")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create k8s rest client: %s", err)
	}

	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	// delete vs
	ic.NetworkingV1alpha3().VirtualServices(namespace).Delete(context.TODO(), newWeights.Svcname, v1.DeleteOptions{})

	var (
		httpRouteList            []*networkingv1alpha3.HTTPRoute
		HTTPRouteDestinationList []*networkingv1alpha3.HTTPRouteDestination
	)

	log.Printf("version: %+v, weight: %+v\n", newWeights.Versions, newWeights.Weights)
	for i := 0; i < len(newWeights.Versions); i++ {
		// 定义http路由
		HTTPRouteDestination := &networkingv1alpha3.HTTPRouteDestination{
			Destination: &networkingv1alpha3.Destination{
				Host:   newWeights.Svcname,
				Subset: newWeights.Versions[i],
			},
			// 定义权重
			Weight: newWeights.Weights[i],
		}
		HTTPRouteDestinationList = append(HTTPRouteDestinationList, HTTPRouteDestination)
	}
	/*
		HTTPRouteDestination1 := &networkingv1alpha3.HTTPRouteDestination{
			Destination: &networkingv1alpha3.Destination{
				Host:   "reviews", //newWeights.Svcname,
				Subset: "v1",      //newWeights.Versions[i],
			},
			// 定义权重
			Weight: 27, //newWeights.Weights[i],
		}
		HTTPRouteDestinationList = append(HTTPRouteDestinationList, HTTPRouteDestination1)

		HTTPRouteDestination2 := &networkingv1alpha3.HTTPRouteDestination{
			Destination: &networkingv1alpha3.Destination{
				Host:   "reviews", //newWeights.Svcname,
				Subset: "v2",      //newWeights.Versions[i],
			},
			// 定义权重
			Weight: 73, //newWeights.Weights[i],
		}
		HTTPRouteDestinationList = append(HTTPRouteDestinationList, HTTPRouteDestination2)
	*/
	httpRouteSign := networkingv1alpha3.HTTPRoute{

		Route: HTTPRouteDestinationList,
	}
	httpRouteList = append(httpRouteList, &httpRouteSign)
	virtualService := &v1alpha3.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      newWeights.Svcname,
			Namespace: namespace,
		},
		Spec: networkingv1alpha3.VirtualService{
			Hosts:    []string{newWeights.Svcname}, // 定义可访问的hosts
			Gateways: []string{"reactapp-gateway"},
			Http:     httpRouteList, // 为hosts 绑定路由
		},
	}
	// 创建VS
	vs, err := ic.NetworkingV1alpha3().VirtualServices(namespace).Create(context.TODO(), virtualService, v1.CreateOptions{})
	if err != nil {
		return
	}
	// 打印VS
	log.Print(vs)

	/*
		//To allocate slice for request body
		length, err := strconv.Atoi(c.Request.Header.Get("Content-Length"))
		if err != nil {
			//c.WriteHeader(http.StatusInternalServerError)
			return
		}

			//Read body data to parse json
			body := make([]byte, length)
			length, err = c.Request.Form.Get("artist") //Body.Read(body)
			if err != nil && err != io.EOF {
				//c.WriteHeader(http.StatusInternalServerError)
				return
			}


				//parse json
				var jsonBody map[string]interface{}
				err = json.Unmarshal(body[:length], &jsonBody)
				if err != nil {
					//c.WriteHeader(http.StatusInternalServerError)
					return
				}
				fmt.Printf("%v\n", jsonBody)
				fmt.Printf("%s\n", jsonBody.artist)

				//fmt.Println(b)
				c.IndentedJSON(http.StatusCreated, jsonBody)
	*/
	//c.IndentedJSON(http.StatusCreated, newWeights)

	c.IndentedJSON(http.StatusCreated, vs)
}

// postLBConfig
// @Summary defines load balance policy that applies to traffic intended for a service after routing has occurred.
// @Tags Istio Resouce Config
// @Accept  json
// @Produce  json
// @Param title body string true "title"
// @Param body body string true "body"
// @Success 201
// @Failure 400
// @Router /api/icg/defineLBConfig [post]
func postLBConfig(c *gin.Context) {

	var newlb lb

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newlb); err != nil {
		return
	}

	log.Printf("posted lb config: %+v\n", newlb)

	namespace := newlb.Ns
	kubeconfig := os.Getenv("KUBECONFIG") // os.GEtenv gets environment variable

	if len(kubeconfig) == 0 || len(namespace) == 0 {
		log.Fatalf("Environment variables KUBECONFIG and NAMESPACE need to be set")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create k8s rest client: %s", err)
	}

	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	var (
		destinationRule *v1alpha3.DestinationRule
		subsetList      []*networkingv1alpha3.Subset
	)

	// 设置subset
	subset := &networkingv1alpha3.Subset{
		Name:   "v2",
		Labels: map[string]string{"version": "v2"},
		//TrafficPolicy:        nil,

	}
	subsetList = append(subsetList, subset)

	destinationRule = &v1alpha3.DestinationRule{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespace,
			Name:      "dr-" + newlb.Svcname,
		},
		Spec: networkingv1alpha3.DestinationRule{
			Host:    newlb.Svcname,
			Subsets: subsetList,
			TrafficPolicy: &networkingv1alpha3.TrafficPolicy{
				LoadBalancer: &networkingv1alpha3.LoadBalancerSettings{
					LbPolicy: &networkingv1alpha3.LoadBalancerSettings_Simple{
						Simple: networkingv1alpha3.LoadBalancerSettings_PASSTHROUGH,
					},
					LocalityLbSetting: nil,
				},
				ConnectionPool: &networkingv1alpha3.ConnectionPoolSettings{
					Tcp: &networkingv1alpha3.ConnectionPoolSettings_TCPSettings{
						// Maximum number of HTTP1 /TCP connections to a destination host. Default 2^32-1.
						MaxConnections: 200,
						// TCP connection timeout. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s.
						ConnectTimeout: nil,
					},
					Http: &networkingv1alpha3.ConnectionPoolSettings_HTTPSettings{
						// Maximum number of pending HTTP requests to a destination. Default 2^32-1.
						// 最大请求数
						Http1MaxPendingRequests: 200,
						// Maximum number of requests to a backend. Default 2^32-1.
						// 每个后端最大请求数
						Http2MaxRequests: 20,
						// Maximum number of requests per connection to a backend. Setting this
						// parameter to 1 disables keep alive. Default 0, meaning "unlimited",
						// up to 2^29.
						// 是否启用keepalive对后端进行长链接 0 表示启用
						MaxRequestsPerConnection: 0,
						// Maximum number of retries that can be outstanding to all hosts in a
						// cluster at a given time. Defaults to 2^32-1.
						// 在给定时间内最大的重试次数
						MaxRetries: 1,
						// The idle timeout for upstream connection pool connections. The idle timeout is defined as the period in which there are no active requests.
						// If not set, the default is 1 hour. When the idle timeout is reached the connection will be closed.
						// Note that request based timeouts mean that HTTP/2 PINGs will not keep the connection alive. Applies to both HTTP1.1 and HTTP2 connections.
						// 不设置默认1小时没有请求，断开后端连接
						IdleTimeout: nil,
						// Specify if http1.1 connection should be upgraded to http2 for the associated destination.
						//H2UpgradePolicy:          0,
					},
				},
				// 类似nginx的 next upstream
				//OutlierDetection:     nil,
				//Tls:                  nil,
				PortLevelSettings: nil,
			},
		},
	}
	dr, err := ic.NetworkingV1alpha3().DestinationRules(namespace).Create(context.TODO(), destinationRule, v1.CreateOptions{})
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(dr)

	c.IndentedJSON(http.StatusCreated, dr)
}

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld")
}
func main() {

	router := gin.Default()
	router.GET("/api/icg/hello", getExample)
	router.GET("/api/icg/istioConfig", getIstioConfig)
	router.POST("/api/icg/weightConfig", postWeightConfig)
	router.POST("/api/icg/postBConfig", postLBConfig)

	// swagger uiを開く
	// http://34.146.130.74:3011/swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run("0.0.0.0:3011")

}
