set -x

minikube start --kubernetes-version v1.13.0

GOOS=linux go build -o ./app .

tag=${1:-$RANDOM}
tagged_build_name=gslc-${tag}
container_name=demo-${tagged_build_name}

eval $(minikube docker-env)

docker build -t ${tagged_build_name} .

kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:default 2>/dev/null

kubectl run --rm -i ${container_name} --image=${tagged_build_name} --image-pull-policy=Never


#kubectl run --rm -i ${container_name} --image=${tagged_build_name} –image-pull-policy=Never --generator=run-pod/v1 -n gslc -v=5


