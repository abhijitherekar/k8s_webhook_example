package main

// the main logic file, this file will have the webhook logic
import (
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	admissionWebhookValidatingannotation = "k8s_webhook_example/validate"
	nameLabel                            = "app/k8s.io/name/k8s_webhook_example"
	projectLabel                         = "app/k8s.io/pro/k8s_webhook_example"
)

var (
	schema       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(schema)
	deserializer = codecs.UniversalDeserializer()
)

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}

	requiredLabels = []string{
		projectLabel,
		nameLabel,
	}
)

type WebhookSvr struct {
	Server *http.Server
}

type admServerParam struct {
	port    int
	tlsCert string
	tlsKey  string
}

//mutates the admission request
func (ws *WebhookSvr) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	fmt.Println("Mutate")
	return nil
}

//Validates the request
// 1. extract the request
// 2. switch depending on the kind of object
// 3. Unmarshal the object into its kind.
// 4. Now get the metadata, which will have the annotation
// 5. Call the function admissionRequired to check if the pod is annotated with false
// 6. If the annotation says false return reponse true
// 7. Else go further and check the labels.
// 8. If all the required labels are present then send allowed, else not allowed.
func (ws *WebhookSvr) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	request := ar.Request

	var (
		metadata *metav1.ObjectMeta
		labels   map[string]string
	)

	fmt.Println("Admission request for:", request.Kind, request.Name, request.Namespace)

	switch request.Kind.Kind {
	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
			glog.Error("Could not unmarshall", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		metadata, labels = &pod.ObjectMeta, pod.Labels
	}
	if !admissionRequired(admissionWebhookValidatingannotation, metadata) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}
	for _, v := range requiredLabels {
		if _, ok := labels[v]; !ok {
			return &v1beta1.AdmissionResponse{
				Allowed: false,
			}
		}
	}
	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

//check if the annonatation says if admission check is
//required or not.
//Checks whether the annote is in the metadata
func admissionRequired(annote string, meta *metav1.ObjectMeta) bool {
	objAnnot := meta.GetAnnotations()

	if objAnnot == nil {
		objAnnot = map[string]string{}
	}

	switch strings.ToLower(objAnnot[annote]) {
	case "false":
		return false
	default:
		return true
	}

}

//Serve func is called as the entry point from for the webhook
// 1. it reads the request body
// 2. Check the header type, k8s-api have JSON objects
//    so make sure
// 2. Unmarshalls body into the v1beta/admissionreview
// 3. depending on the request send to mutate func or the validate function
// 4. Gets the response and sends it back to the api-server
func (ws *WebhookSvr) Serve(w http.ResponseWriter, r *http.Request) {
	var data []byte
	if _, err := ioutil.ReadAll(r.Body); err != nil {
		log.Println("failed to read the body")
		http.Error(w, "read body", http.StatusBadRequest)
	}

	if len(data) == 0 {
		glog.Error("request data is nil")
		http.Error(w, "empty body", http.StatusBadRequest)
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid content expect json", http.StatusBadRequest)

	}

	var admResp *v1beta1.AdmissionResponse
	admReview := v1beta1.AdmissionReview{}

	if _, _, err := deserializer.Decode(data, nil, &admReview); err != nil {
		glog.Error("error desearlizing the object")
		admResp = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if r.URL.Path == "/mutate" {
			admResp = ws.mutate(&admReview)
		} else if r.URL.Path == "/validate" {
			admResp = ws.validate(&admReview)
		}
	}
	doneReview := v1beta1.AdmissionReview{}
	if admResp != nil {
		doneReview.Response = admResp
		if admReview.Request != nil {
			doneReview.Response.UID = admReview.Request.UID
		}
	}

	resp, err := json.Marshal(doneReview)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(resp); err != nil {
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
}
