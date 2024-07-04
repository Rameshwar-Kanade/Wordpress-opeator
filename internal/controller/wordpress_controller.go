/*
Copyright 2024 Rameshwar Kanade.

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

	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	webappv1alpha1 "github.com/Rameshwar-Kanade/wordpress/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WordpressReconciler reconciles a Wordpress object
type WordpressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.rameshwarkanade.online,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.rameshwarkanade.online,resources=wordpresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.rameshwarkanade.online,resources=wordpresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wordpress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *WordpressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Wordpress instance
	wordpress := &webappv1alpha1.Wordpress{}
	err := r.Get(ctx, req.NamespacedName, wordpress)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Wordpress resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Wordpress")
		return ctrl.Result{}, err
	}

	log.Info("Fetched Wordpress resource", "Wordpress", wordpress)

	// Define the desired state
	desiredConfigMap := r.createConfigMap(ctx, wordpress)
	log.Info("Desired ConfigMap created", "ConfigMap", desiredConfigMap)

	desiredSecret := r.createSecret(ctx, wordpress)
	log.Info("Desired Secret created", "Secret", desiredSecret)

	desiredDeployment := r.createDeployment(ctx, wordpress)
	log.Info("Desired Deployment created", "Deployment", desiredDeployment)

	desiredHPA := r.createHPA(ctx, wordpress)
	log.Info("Desired HPA created", "HPA", desiredHPA)

	// Check if the ConfigMap already exists
	existingConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: desiredConfigMap.Name, Namespace: desiredConfigMap.Namespace}, existingConfigMap)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating ConfigMap", "ConfigMap.Namespace", desiredConfigMap.Namespace, "ConfigMap.Name", desiredConfigMap.Name)
		err = r.Create(ctx, desiredConfigMap)
		if err != nil {
			log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", desiredConfigMap.Namespace, "ConfigMap.Name", desiredConfigMap.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	// Check if the Secret already exists
	existingSecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: desiredSecret.Name, Namespace: desiredSecret.Namespace}, existingSecret)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Secret", "Secret.Namespace", desiredSecret.Namespace, "Secret.Name", desiredSecret.Name)
		err = r.Create(ctx, desiredSecret)
		if err != nil {
			log.Error(err, "Failed to create new Secret", "Secret.Namespace", desiredSecret.Namespace, "Secret.Name", desiredSecret.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Secret")
		return ctrl.Result{}, err
	}

	// Check if the Deployment already exists
	existingDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, existingDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "Deployment.Namespace", desiredDeployment.Namespace, "Deployment.Name", desiredDeployment.Name)
		err = r.Create(ctx, desiredDeployment)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", desiredDeployment.Namespace, "Deployment.Name", desiredDeployment.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Update the Deployment if necessary
	if !reflect.DeepEqual(desiredDeployment.Spec, existingDeployment.Spec) {
		log.Info("Updating Deployment", "Deployment.Namespace", existingDeployment.Namespace, "Deployment.Name", existingDeployment.Name)
		err = r.Update(ctx, desiredDeployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", existingDeployment.Namespace, "Deployment.Name", existingDeployment.Name)
			return ctrl.Result{}, err
		}
	}

	// Check if the HPA already exists
	existingHPA := &autoscalingv1.HorizontalPodAutoscaler{}
	err = r.Get(ctx, types.NamespacedName{Name: desiredHPA.Name, Namespace: desiredHPA.Namespace}, existingHPA)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating HPA", "HPA.Namespace", desiredHPA.Namespace, "HPA.Name", desiredHPA.Name)
		err = r.Create(ctx, desiredHPA)
		if err != nil {
			log.Error(err, "Failed to create new HPA", "HPA.Namespace", desiredHPA.Namespace, "HPA.Name", desiredHPA.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get HPA")
		return ctrl.Result{}, err
	}

	// Update the HPA if necessary
	if !reflect.DeepEqual(desiredHPA.Spec, existingHPA.Spec) {
		log.Info("Updating HPA", "HPA.Namespace", existingHPA.Namespace, "HPA.Name", existingHPA.Name)
		err = r.Update(ctx, desiredHPA)
		if err != nil {
			log.Error(err, "Failed to update HPA", "HPA.Namespace", existingHPA.Namespace, "HPA.Name", existingHPA.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *WordpressReconciler) createConfigMap(ctx context.Context, wordpress *webappv1alpha1.Wordpress) *corev1.ConfigMap {
	log := log.FromContext(ctx)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name + "-config",
			Namespace: wordpress.Namespace,
		},
		Data: map[string]string{
			"configData": wordpress.Spec.ConfigData,
		},
	}
	log.Info("Creating ConfigMap", "ConfigMap", configMap)
	return configMap
}

func (r *WordpressReconciler) createSecret(ctx context.Context, wordpress *webappv1alpha1.Wordpress) *corev1.Secret {
	log := log.FromContext(ctx)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name + "-secret",
			Namespace: wordpress.Namespace,
		},
		Data: map[string][]byte{
			"dbUsername": []byte(wordpress.Spec.DbUsername),
			"dbPassword": []byte(wordpress.Spec.DbPassword),
		},
	}
	log.Info("Creating Secret", "Secret", secret)
	return secret
}

func (r *WordpressReconciler) createDeployment(ctx context.Context, wordpress *webappv1alpha1.Wordpress) *appsv1.Deployment {
	log := log.FromContext(ctx)
	replicas := wordpress.Spec.Replicas
	labels := map[string]string{
		"app": wordpress.Name,
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name,
			Namespace: wordpress.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "wordpress",
						Image: wordpress.Spec.Image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "http",
						}},
						Env: []corev1.EnvVar{
							{
								Name: "WORDPRESS_DB_HOST",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: wordpress.Name + "-secret",
										},
										Key: "dbUsername",
									},
								},
							},
							{
								Name: "WORDPRESS_DB_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: wordpress.Name + "-secret",
										},
										Key: "dbPassword",
									},
								},
							},
						},
					}},
				},
			},
		},
	}
	log.Info("Creating Deployment", "Deployment", deployment)
	return deployment
}

func (r *WordpressReconciler) createHPA(ctx context.Context, wordpress *webappv1alpha1.Wordpress) *autoscalingv1.HorizontalPodAutoscaler {
	log := log.FromContext(ctx)
	hpa := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name + "-hpa",
			Namespace: wordpress.Namespace,
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       wordpress.Name,
			},
			MinReplicas:                    &wordpress.Spec.MinReplicas,
			MaxReplicas:                    wordpress.Spec.MaxReplicas,
			TargetCPUUtilizationPercentage: &wordpress.Spec.TargetCPUUtilizationPercentage,
		},
	}
	log.Info("Creating HPA", "HPA", hpa)
	return hpa
}

// SetupWithManager sets up the controller with the Manager.
func (r *WordpressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1alpha1.Wordpress{}).
		Complete(r)
}
