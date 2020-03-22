package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"

	"github.com/ghodss/yaml"

	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	log "github.com/sirupsen/logrus"

	"github.com/ericchiang/k8s"
)

func makeKubeconfigClient(path string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(k8s.Config)
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	client, err := k8s.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func makeClient() (*k8s.Client, error) {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return makeKubeconfigClient(kubeconfig)
	}
	return k8s.NewInClusterClient()
}

func main() {

	kubernetesClient, err := makeClient()

	if err != nil {
		log.WithField("err", err).Error("Could not create kubernetes client")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go watchSecrets(ctx, kubernetesClient)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan

	log.Info("Terminating")
	cancel()
}

func watchSecrets(ctx context.Context, client *k8s.Client) {
	for {
		log.Trace("start secret watch")
		watchSecret := &corev1.Secret{}
		nodeWatcher, err := client.Watch(ctx, k8s.AllNamespaces, watchSecret)
		if err != nil {
			log.WithField("err", err).Fatal("secret watch failed, rbac ok?")
			break
		}

		for {
			secret := &corev1.Secret{}
			t, err := nodeWatcher.Next(secret)
			if err != nil {
				if !strings.Contains(err.Error(), "EOF") {
					log.WithField("err", err).Fatal("Watch errored")
				} else {
					log.Debug("Restarting Watch, API server ended our watch")
					break
				}
			}

			if t == k8s.EventDeleted {
				continue
			}

			if *secret.Type != "kubernetes.io/tls" {
				continue
			}

			log.WithFields(log.Fields{
				"name":      *secret.GetMetadata().Name,
				"namespace": *secret.GetMetadata().Namespace,
			}).Info("Working at Resource")
			caCrt, isCaCrtOk := secret.Data["ca.crt"]
			tlsCa, isTLSCaOk := secret.Data["tls.ca"]

			if !isCaCrtOk {
				log.WithFields(log.Fields{
					"name":      *secret.GetMetadata().Name,
					"namespace": *secret.GetMetadata().Namespace,
				}).Info("Certificate does not contain a ca.crt entry.")
				continue
			}

			if !isTLSCaOk || !bytes.Equal(caCrt, tlsCa) {
				log.WithFields(log.Fields{
					"name":      *secret.GetMetadata().Name,
					"namespace": *secret.GetMetadata().Namespace,
				}).Info("Updating Resource with tls.ca")
				secret.Data["tls.ca"] = caCrt
				client.Update(ctx, secret)
			}

		}
	}
}
