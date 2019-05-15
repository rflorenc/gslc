apiserver=
cluster=
user=
ca=
client_cert=
private_key=

outkubecfg=admin.kubeconfig

kubectl config set-cluster ${cluster} \
	--certificate-authority=${ca} \
	--embed-certs=true \
	--server=${apiserver} \
	--kubeconfig=${outkubecfg}

kubectl config set-credentials ${user} \
	--client-certificate=${client_cert} \
	--client-key=${private_key} \
	--embed-certs=true \
	--kubeconfig=${outkubecfg}

kubectl config set-context default \
	--cluster=${cluster} \
	--user=${user} \
	--kubeconfig=${outkubecfg}

kubectl config use-context default --kubeconfig=${outkubecfg}
cp -v ${outkubecfg} $HOME/.kube/


