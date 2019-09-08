package serverlessorchestrationapp

import (
	"context"
	appv1alpha1 "github.com/kiegroup/serverless-orchestration-operator/pkg/apis/app/v1alpha1"
	oappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_serverlessorchestrationapp")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ServerlessOrchestrationApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileServerlessOrchestrationApp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("serverlessorchestrationapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ServerlessOrchestrationApp
	err = c.Watch(&source.Kind{Type: &appv1alpha1.ServerlessOrchestrationApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchOwnedObjects := []runtime.Object{
		&oappsv1.DeploymentConfig{},
		&corev1.Service{},
		&routev1.Route{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.ServerlessOrchestrationApp{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileServerlessOrchestrationApp implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileServerlessOrchestrationApp{}

// ReconcileServerlessOrchestrationApp reconciles a ServerlessOrchestrationApp object
type ReconcileServerlessOrchestrationApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ServerlessOrchestrationApp object and makes changes based on the state read
// and what is in the ServerlessOrchestrationApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServerlessOrchestrationApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ServerlessOrchestrationApp")

	// Fetch the ServerlessOrchestrationApp instance
	instance := &appv1alpha1.ServerlessOrchestrationApp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a CM
	cm := newCMForCR(instance)

	// Set ServerlessOrchestrationApp instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, cm, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this CM already exists
	foundCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundCM)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "CM.Namespace", cm.Namespace, "CM.Name", cm.Name)
		err = r.client.Create(context.TODO(), cm)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("ConfigMap created successfully")
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a DC
	dc := newDCForCR(instance, cm)

	// Set ServerlessOrchestrationApp instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, dc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DC already exists
	foundDC := &oappsv1.DeploymentConfig{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dc.Name, Namespace: dc.Namespace}, foundDC)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DC", "DC.Namespace", dc.Namespace, "DC.Name", dc.Name)
		err = r.client.Create(context.TODO(), dc)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("DC created successfully")
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a Service
	service := newServiceForCR(dc)

	// Set ServerlessOrchestrationApp instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("Service created successfully")
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a Route
	route := newRouteForCR(service)

	// Set ServerlessOrchestrationApp instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, route, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Route already exists
	foundRoute := &routev1.Route{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, foundRoute)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Route", "Route.Namespace", route.Namespace, "route.Name", route.Name)
		err = r.client.Create(context.TODO(), route)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("Route created successfully")
	} else if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

const (
	workflowJsonKey = "workflow-json"
)

// Create deployment config for the app
func newCMForCR(cr *appv1alpha1.ServerlessOrchestrationApp) *corev1.ConfigMap {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workflow-config",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			workflowJsonKey: cr.Spec.Definition,
		},
	}
}

// Create deployment config for the app
func newDCForCR(cr *appv1alpha1.ServerlessOrchestrationApp, cm *corev1.ConfigMap) *oappsv1.DeploymentConfig {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: oappsv1.DeploymentConfigSpec{
			Strategy: oappsv1.DeploymentStrategy{
				Type: oappsv1.DeploymentStrategyTypeRolling,
			},
			Replicas: 1,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Spec.Name,
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            cr.Spec.Name,
							Image:           cr.Spec.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports:           cr.Spec.Ports,
							Env: []corev1.EnvVar{
								{
									Name:  "WORKFLOW_PATH",
									Value: "/opt/process/workflow.json",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workflow-json",
									MountPath: "/opt/process",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workflow-json",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cm.Name,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  workflowJsonKey,
											Path: "workflow.json",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Create service for the app
func newServiceForCR(dc *oappsv1.DeploymentConfig) *corev1.Service {
	ports := []corev1.ServicePort{}
	for _, port := range dc.Spec.Template.Spec.Containers[0].Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.ContainerPort,
			TargetPort: intstr.FromInt(int(port.ContainerPort)),
		},
		)
	}

	return &corev1.Service{
		ObjectMeta: dc.ObjectMeta,
		Spec: corev1.ServiceSpec{
			Selector: dc.Spec.Selector,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
		},
	}
}

// Create droute for the app
func newRouteForCR(service *corev1.Service) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: service.ObjectMeta,
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: service.Name,
			},
		},
	}
}
