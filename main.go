package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	versionedclient "istio.io/client-go/pkg/clientset/versioned"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// swagger UI
	// docsのディレクトリを指定
	_ "web-service-gin/docs" // ←追記

	ginSwagger "github.com/swaggo/gin-swagger"   // ←追記
	"github.com/swaggo/gin-swagger/swaggerFiles" // ←追記
)

// ac "istio.io/client-go/pkg/applyconfiguration"
// istio.io/client-go/pkg/applyconfiguration/utils.go
// /Users/satoshiaikawa/client-go-master/pkg/applyconfiguration/utils.go

// weights represents data about a record album.
type weights struct {
	Ns       string   `json:"ns"`
	Svcname  string   `json:"svcname"`
	Versions []string `json:"versions"`
	Weights  []int32  `json:"weights"`
}

// getAlbums responds with the list of all albums as JSON.
func getExample(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "Hello Istio Client Go")
}

// getIstioConfig responds with the list of all as JSON.
// @Summary lists istio configurations of intio-gateway, virtual service, and destination rules.
// @Tags Todo
// @Produce  json
// @Success 200
// @Failure 400
// @Router /api/icg/istioConfig [get]
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
	vsList, err := ic.NetworkingV1alpha3().VirtualServices(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get VirtualService in %s namespace: %s", namespace, err)
	}

	for i := range vsList.Items {
		vs := vsList.Items[i]
		log.Printf("Index: %d VirtualService Hosts: %+v\n", i, vs.Spec.GetHosts())
		log.Printf("%+v\n", vs.Spec.GetHttp()[0].GetRoute()[0].GetWeight())
	}

	// Test DestinationRules
	drList, err := ic.NetworkingV1alpha3().DestinationRules(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get DestinationRule in %s namespace: %s", namespace, err)
	}

	for i := range drList.Items {
		dr := drList.Items[i]
		log.Printf("Index: %d DestinationRule Host: %+v\n", i, dr.Spec.GetHost())
	}

	// Test Gateway
	gwList, err := ic.NetworkingV1alpha3().Gateways(namespace).List(context.TODO(), metav1.ListOptions{})
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
	seList, err := ic.NetworkingV1alpha3().ServiceEntries(namespace).List(context.TODO(), metav1.ListOptions{})
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
// @Tags Todo
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

	// swagger uiを開く
	// http://34.146.130.74:3011/swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run("0.0.0.0:3011")

}
