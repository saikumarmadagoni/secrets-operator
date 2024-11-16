/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"fmt"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	mychartv1 "github.com/saikumarmadagoni/secrets-operator/api/v1"

	//	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var logger = log.Log.WithName("controller_scaler")

// AwssecretReconciler reconciles a Awssecret object
type AwssecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mychart.my-chart.io,resources=awssecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mychart.my-chart.io,resources=awssecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mychart.my-chart.io,resources=awssecrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Awssecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *AwssecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	log := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)

	log.Info("Reconcile called")

	// TODO(user): your logic here
	secretCrd := &mychartv1.Awssecret{}

	err := r.Get(ctx, req.NamespacedName, secretCrd)

	if err != nil {

		if apierrors.IsNotFound(err) {
			log.Info("secretcrd resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed")
		return ctrl.Result{}, err
	}

	awsSecret := secretCrd.Spec.AwsSecretName

	fmt.Println("the aws secret", awsSecret)

	awssecretkeys := secretCrd.Spec.AwsSecretKeys

	kubernetessecret := secretCrd.Spec.KubernetesSecretName

	namespace := req.Namespace

	fmt.Println("crd params", awssecretkeys, kubernetessecret, namespace)

	secretName := awsSecret
	region := "us-east-1"

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))

	fmt.Println("creds added")

	if err != nil {
		fmt.Println(err)
	}

	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)

	if err != nil {
		// For a list of exceptions thrown, see
		// https://<<{{DocsDomain}}>>/secretsmanager/latest/apireference/API_GetSecretValue.html
		fmt.Println(err.Error())
	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	fmt.Printf("type of secret string %T", secretString)

	var mapresult map[string]string

	// Unmarshal the JSON string into the map
	error := json.Unmarshal([]byte(secretString), &mapresult)
	if error != nil {
		fmt.Println("Error unmarshalling JSON:", err)

	}

	// Print the map to verify the result
	fmt.Println(mapresult)

	var dataobject = make(map[string][]byte)

	for key,value := range mapresult {
		dataobject[key]=[]byte(value)
	}

	// var byteusername = []byte(mapresult["username"])

	// var bytepassword = []byte(mapresult["password"])
	

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubernetessecret, // Name of the secret
			Namespace: req.Namespace,    // Namespace in which to create the secret
		},
		Data: dataobject,
	}

	existingSecret := &corev1.Secret{}

	errorget := r.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}, existingSecret)
	if errorget == nil {
		// Secret already exists, no need to create it
		fmt.Println("Secret already exists need to update if any changes ")
		
	}
	/* else if client.IgnoreNotFound(err) == nil {
		fmt.Println("secret do not exist error")
		return ctrl.Result{}, err
	}

	// Set the owner reference to enable garbage collection when the CR is deleted
	owner := &mychartv1.Awssecret{} // Replace with your actual CR type
	if err := r.Client.Get(ctx, req.NamespacedName, owner); err == nil {
		if err := controllerutil.SetControllerReference(owner, secret, r.Scheme); err != nil {
			fmt.Println("client get error")
			return ctrl.Result{}, err
		}
	}
	*/
	// Create the secret in the cluster
	if err := r.Client.Create(ctx, secret); err != nil {
		fmt.Printf("Failed to create secret: %v\n", err)
		return ctrl.Result{}, err
	}

	fmt.Println("Secret created successfully")
	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *AwssecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mychartv1.Awssecret{}).
		Complete(r)
}
