#!/usr/bin/env groovy
pipeline {
    parameters {
        booleanParam (name: 'K8S_DEPLOY', defaultValue: false,     description: 'Whether to keep deployed pod on k8s')
        booleanParam (name: 'KEEP_ARTIFACTS', defaultValue: false,     description: 'Whether to keep produced artifacts in the repo')
        booleanParam (name: 'DEBUG_IMAGES', defaultValue: false, description: 'Whether to build debug images')

        string (name: 'IMAGE_NAME', defaultValue: 'goidp', description: 'Image name')

        string (name: 'DOCKER_REG', defaultValue: 'docker.io', description: 'Docker registry')
        string (name: 'DOCKER_CRED', defaultValue: 'docker-registry-credentials', description: 'Docker registry credentials ID')

        string (name: 'GIT_CRED', defaultValue: 'git-credentials', description: 'Git credentials to pass in the docker build for bitbucket address')

        string (name: 'IMG_PULL_SECRET_CRED', defaultValue: 'image-pull-secret', description: 'The Kubernetes secret for the Docker registry (imagePullSecrets)')

        string (name: 'K8S_CRED', defaultValue: 'kube-config', description: 'Kubernetes kube config credentials ID')

        string (name: 'HELM_REPO_BASE', defaultValue: 'https://repo.helm.com', description: 'Your helm repository')
        string (name: 'HELM_CRED', defaultValue: 'helm-repo-credentials', description: 'Helm repository credentials ID')

        string (name: 'AGENT_LABEL', defaultValue: 'docker-worker', description: 'Jenkins agent label name')

        string (name: 'API_VERSION', defaultValue: 'v1.0', description: 'Api version')
    }
    environment {
        DOCKER = credentials("${params.DOCKER_CRED}")
        HELM = credentials("${params.HELM_CRED}")
        GIT = credentials("${params.GIT_CRED}")

        DEV_NAMESPACE = "${params.IMAGE_NAME}-${env.GIT_COMMIT.take(8)}"
        APP_BUILD = "${env.BUILD_ID}"
        API_VERSION = "${params.API_VERSION}"
    }
   agent { label "${params.AGENT_LABEL}" }
   stages {
        stage('Setup') {
            steps {
                script {
                    env.APP_VERSION = getAppVersion()
                    env.APP_VERSION_SHORT = getAppVersionShort("${APP_VERSION}")
                    env.CHART_VERSION = getChartVersion("${APP_VERSION_SHORT}")
                    env.RELEASE_NAME = validateHelmReleaseName("${IMAGE_NAME}-${APP_VERSION}") // how it is deployed on k8s
                    env.CHART_NAME = "${IMAGE_NAME}-${CHART_VERSION}"
                    env.REGISTRY = "${DOCKER_REG}/"
                    env.REPOSITORY = generateDockerUrls(BRANCH_NAME, APP_VERSION_SHORT)
                    env.DOCKER_REPO = "https://${DOCKER_REG}:443/artifactory/" + generateDockerUrls(BRANCH_NAME, APP_VERSION_SHORT)
                    env.HELM_REPO = "${params.HELM_REPO_BASE}"
                    env.JTEST_POD_NAME = "test${env.GIT_COMMIT.take(8)}"
                    withCredentials([file(credentialsId: "${IMG_PULL_SECRET_CRED}", variable: 'SECRET_FILE')]) {
                        def IMG_PULL_SECRET_FILE = readYaml file: "${SECRET_FILE}"
                        env.IMG_PULL_SECRET_NAME= IMG_PULL_SECRET_FILE.metadata.name.trim()
                    }
                    println("""
                    Registry pull secret name: ${IMG_PULL_SECRET_NAME}
                    Base version: ${APP_VERSION_SHORT}
                    Docker repo: ${DOCKER_REPO}
                    Helm repo: ${HELM_REPO}
                    Chart name set to: ${CHART_NAME}
                    Chart release name set to: ${RELEASE_NAME}
                    Chart version set to: ${CHART_VERSION}
                    Chart appVersion set to: ${APP_VERSION}
                    Docker image set to: ${IMAGE_NAME}:${APP_VERSION}
                    Test namespace: ${DEV_NAMESPACE}
                    """)
                }
            }
        }
        stage('Docker test') {
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY}", DOCKER_CRED) {
                        BASE = docker.build("${REGISTRY}${IMAGE_NAME}:${APP_VERSION}-test",
                                            "--target=test .")
                    }
                }
            }
        }
        stage('Docker build') {
            steps {
                script {
                    docker.withRegistry("https://${REGISTRY}", DOCKER_CRED) {
                        BASE = docker.build("${REGISTRY}${IMAGE_NAME}:${APP_VERSION}",
                                            "--build-arg APP_BUILD=${APP_BUILD}\
                                             --build-arg APP_NAME=${IMAGE_NAME}\
                                             --build-arg CHART_VERSION=${CHART_VERSION}\
                                             --build-arg API_VERSION=${API_VERSION}\
                                             --build-arg APP_VERSION=${APP_VERSION}\
                                             --no-cache .")
                        BASE.push()
                        if (env.BRANCH_NAME == "master") {
                            BASE.push("latest")
                            BASE.push(APP_VERSION_SHORT)
                        }
                    }
                }
            }
        }
        stage('Docker debug build') {
            when {
                allOf{
                    expression{env.BRANCH_NAME != 'master'}
                    expression{env.BRANCH_NAME.split("/")[0] != 'release'}
                    environment name: 'DEBUG_IMAGES', value: 'true'
                }
            }
            steps {
                script {
                    DEBUG_REGISTRY = "${DOCKER_REG}/giacomocortesi/"
                    docker.withRegistry("https://${DEBUG_REGISTRY}", DOCKER_CRED) {
                        DEBUG_BASE = docker.build("${DEBUG_REGISTRY}${IMAGE_NAME}:${APP_VERSION}-debug",
                                            "-f Dockerfile.debug\
                                             --build-arg APP_BUILD=${APP_BUILD} \
                                             --build-arg APP_NAME=${IMAGE_NAME} \
                                             --build-arg CHART_VERSION=${CHART_VERSION} \
                                             --build-arg API_VERSION=${API_VERSION} \
                                             --build-arg APP_VERSION=${APP_VERSION} .")
                        DEBUG_BASE.push()
                    }
                }
            }
        }
        stage('HELM Linting'){
            steps{
                script{
                    sh 'helm dep up ./helm'
                    sh 'helm lint ./helm'
                }
            }
        }
        stage('HELM templating and packaging'){
            steps{
                script{
                    // Read Chart.yaml and modify appVersion, version and name
                    def CHART_FILE = readYaml file: "${WORKSPACE}/helm/Chart.yaml"
                    CHART_FILE = ['appVersion': APP_VERSION, 'version': CHART_VERSION, 'name': IMAGE_NAME]
                    // Write the modified  Chart.yaml file before packaging
                    writeYaml file:"${WORKSPACE}/helm/Chart.yaml", data: CHART_FILE, overwrite: true

                    // Read Chart.yaml and modify image and tag
                    def CHART_VALUES_FILE = readYaml file: "${WORKSPACE}/helm/values.yaml"
                    CHART_VALUES_FILE.image.registry = "${DOCKER_REG}"
                    CHART_VALUES_FILE.image.repository = "${REPOSITORY}${IMAGE_NAME}"
                    CHART_VALUES_FILE.image.tag = APP_VERSION

                    // Write the modified values.yaml file before packaging
                    writeYaml file:"${WORKSPACE}/helm/values.yaml", data: CHART_VALUES_FILE, overwrite: true

                    // Redirect helm template to temporary file to avoid excessive conole output.
                    sh "helm template ${RELEASE_NAME} ./helm > test.yaml"

                    // Templating just service file and extract the service name and port
                    // to use it during the test phase
                    sh "cd helm && helm template ${RELEASE_NAME} -s templates/service.yaml . > ../template.yaml "
                    def SERVICE_TEMPLATE =  readYaml file: "./template.yaml"
                    SERVICE_TEST_URL = getServiceName(SERVICE_TEMPLATE)

                    // Create helm package and push it to artifactory
                    sh "helm package ./helm"
                    sh "curl -u ${HELM_USR}:${HELM_PSW} -X PUT '${HELM_REPO}${CHART_NAME}.tgz' -T *.tgz"
                }
            }
        }
        stage('Check K8S reachability') {
            steps {
                script {
                    withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
                        def KUBEFILE = readYaml file: "${SECRET_FILE}"
                        def K8S_SERVER = KUBEFILE.clusters[0].cluster.server
                        sh """curl -k  ${K8S_SERVER} -o /dev/null -s -w "Response Code: %{http_code}" """
                    }
                }
            }
        }
        stage('K8S deploy') {
            steps{
                createTestNamespace("${DEV_NAMESPACE}")
                installImagePullSecret("${DEV_NAMESPACE}")
                generateSecret("${DEV_NAMESPACE}")
                helmInstall("${DEV_NAMESPACE}", "${RELEASE_NAME}")
            }
        }
        stage('API tests') {
            steps {
                withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
                    sh 'kubectl --kubeconfig ${SECRET_FILE} -n ${DEV_NAMESPACE} run ${JTEST_POD_NAME} --image yauritux/busybox-curl -- sh -c "tail -f /dev/null"'
                    sh '''
                        while ! $(kubectl --kubeconfig ${SECRET_FILE} -n ${DEV_NAMESPACE} get pod ${JTEST_POD_NAME} -o jsonpath={.status.containerStatuses[*].ready});
                                do echo "Waiting for testing pod to be up & running";
                                done'''
                }
                curlTest ("${DEV_NAMESPACE}", SERVICE_TEST_URL)
            }
        }
        stage('Cleanup k8s') {
            when {
                allOf{
                    environment name: 'K8S_DEPLOY', value: 'false'
                }
            }
            steps {
                // Remove release if exists
                helmDelete ("${DEV_NAMESPACE}", "${RELEASE_NAME}")
                deleteNamespace("${DEV_NAMESPACE}")
            }
        }
        stage('Cleanup artifacts') {
            when {
                allOf{
                    expression{env.BRANCH_NAME != 'master'}
                    expression{env.BRANCH_NAME.split("/")[0] != 'release'}
                    environment name: 'KEEP_ARTIFACTS', value: 'false'
                }
            }
            steps {
                sh "curl -u ${HELM_USR}:${HELM_PSW} -X DELETE '${HELM_REPO}${CHART_NAME}.tgz'"
                sh "curl -u ${HELM_USR}:${HELM_PSW} -X DELETE '${DOCKER_REPO}${IMAGE_NAME}/${APP_VERSION}'"
            }
        }
    }
    post {
        failure {
            script{
                sh 'echo "FAILURE"'
                withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
                    sh 'kubectl --kubeconfig ${SECRET_FILE} -n ${DEV_NAMESPACE} delete pod ${JTEST_POD_NAME}'
                }

                // Uninstall helm-charts and cleanup the namespace just if not a release branch and K8S_DEPLOY=false
                if (K8S_DEPLOY == false || (BRANCH_NAME.split("/")[0] != "release" && BRANCH_NAME != "master" ) ){
                    sh 'echo "Removing helm-chart and namespace" '
                    helmDelete ("${DEV_NAMESPACE}", "${RELEASE_NAME}")
                    deleteNamespace("${DEV_NAMESPACE}")
                }
                else{
                    println("Skipping k8s cleanup")
                }
                // Delete artifacts just if the branch is not a release branch and the var KEEP_ARTIFACTS = false
                if (KEEP_ARTIFACTS == false || (BRANCH_NAME.split("/")[0] != "release" && BRANCH_NAME != "master" ) ){
                    sh 'echo "Deleting artefacts" '
                    sh "curl -u ${HELM_USR}:${HELM_PSW} -X DELETE '${HELM_REPO}${CHART_NAME}.tgz'"
                    sh "curl -u ${HELM_USR}:${HELM_PSW} -X DELETE '${DOCKER_REPO}${IMAGE_NAME}/${APP_VERSION}'"
                }
                else{
                    println("Skipping artefacts cleanup")
                }
            }
        }
        always {
            script {
                currentBuild.result = currentBuild.result ?: 'SUCCESS'
            }
        }
    }
}

// getAppVersion retrieve the app version
def getAppVersion() {
    // Check branch type
    BRANCH_TYPE = env.BRANCH_NAME.split("/")[0]
    if (BRANCH_TYPE == "master" || BRANCH_TYPE == "develop" || BRANCH_TYPE == "release") {
        return sh(script: "git describe --tags --always --first-parent --long --match 'ver[0-9]*' | sed -e 's|^ver||' -e 's|-|.|g'", returnStdout: true).trim().toLowerCase()
    }
    def JIRA_REF = (env.BRANCH_NAME =~ /(?<=\/)(.*?)[^-]*-[^-]*/)
    return (JIRA_REF) ? "${JIRA_REF.group(0)}-${env.BUILD_ID}".toLowerCase() : "${BRANCH_NAME}-${env.BUILD_ID}".toLowerCase()
}

// getAppVersionShort retrieve the shorted app version
def getAppVersionShort(appVersion) {
    return sh(script: "echo ${appVersion} | cut -f 1,2,3 -d '.'", returnStdout: true).trim().toLowerCase()
}

// getChartVersion returns chart version
def getChartVersion(baseVersion) {
    BRANCH_TYPE= env.BRANCH_NAME.split("/")[0]
    if (BRANCH_TYPE == "master" || BRANCH_TYPE == "develop" || BRANCH_TYPE == "release") {
        CHART_VERSION="${baseVersion}"
    } else {
        CHART_VERSION="0.0.${env.BUILD_ID}"
    }
    return CHART_VERSION
}

def validateHelmReleaseName (releaseName) {
    // Helm release name has to be all lowercase and long 53 characters at maximum
    return releaseName.take(53).toLowerCase().replaceAll("\\.", "-")
}

def helmInstall (namespace, release) {
    echo "Installing ${release} in ${namespace}"
    helmDelete(namespace, release)
    withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
        sh """
               helm install --namespace ${namespace} --kubeconfig ${SECRET_FILE} ${release} ./helm  \
                --set image.repository=${REPOSITORY}${IMAGE_NAME} \
                --set jwt.publicKeySecretKey=public.pem \
                --set jwt.privateKeySecretKey=private.pem \
                --set global.imagePullSecrets[0]=${IMG_PULL_SECRET_NAME} \
                --set service.type=NodePort \
                --atomic --timeout 300s
        """
    }
}

def helmDelete (namespace, release) {
    echo "Deleting ${release} in ${namespace} if deployed"
    withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
        sh "[ -z \"\$(helm list --namespace ${namespace} --kubeconfig ${SECRET_FILE} --short --filter ${release} 2>/dev/null)\" ] || helm delete --kubeconfig ${SECRET_FILE} --namespace ${namespace} ${release}"
    }
}

// generateSecret creates a kubernetes secret that the identity provider uses for JWT management
def generateSecret(namespace) {
    sh "openssl genrsa -out private.pem -passout pass:keysecret 2048 1>/dev/null 2>&1"
    sh "openssl rsa -in private.pem -out public.pem -RSAPublicKey_out 1>/dev/null 2>&1"
    withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
        // delete idp-password secret if present (it is created by default if missing)
        sh "[ -z \"\$(kubectl --namespace ${namespace} --kubeconfig ${SECRET_FILE} get secret idp-password 2>/dev/null)\" ] || kubectl --namespace ${namespace} delete secret idp-password --kubeconfig=${SECRET_FILE}"
        sh "kubectl --namespace ${namespace} --dry-run=client -o yaml create secret generic goidp-rsa --kubeconfig=${SECRET_FILE} --from-file=private.pem=private.pem --from-file=public.pem=public.pem | kubectl --kubeconfig ${SECRET_FILE} -n ${namespace} apply -f -"
    }
}

// Test with a simple curl and check we get 200 back
def curlTest (namespace, test_url) {
    echo "Running tests in ${namespace}"
    withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
        retry(5) {
            timeout(time: 15, unit: 'SECONDS') {
                def result = sh (
                        returnStdout: true,
                        script: "kubectl --kubeconfig ${SECRET_FILE} -n ${namespace} exec ${JTEST_POD_NAME} -- curl -H \"Content-Type: application/vnd.api+json\" -o /dev/null -s -w \"Response Code: %{http_code} \nTotal execution time: %{time_total}\nTotal download size: %{size_download}\n\" ${test_url}"
                    )
                    println("Response testing endpoint: ${test_url}: \n${result}")
                }
            }
    }
}

//  Create the kubernetes namespace
def createTestNamespace (namespace){
    echo "Creating namespace ${namespace} if needed"
    withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
        sh "[ ! -z \"\$(kubectl get ns --kubeconfig=${SECRET_FILE} ${namespace} -o name 2>/dev/null)\" ] || kubectl --kubeconfig=${SECRET_FILE} create ns ${namespace}"
    }
}

// Delete the kubernetes namespace
def deleteNamespace (namespace){
    echo "Deleting namespace ${namespace} if needed"
    withCredentials([file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
        sh "[ ! -z \"\$(kubectl get ns --kubeconfig=${SECRET_FILE} ${namespace} -o name 2>/dev/null)\" ] && kubectl --kubeconfig=${SECRET_FILE} delete ns ${namespace}"
    }
}

// Install the image pull secret in the target namespace
def installImagePullSecret (namespace){
    withCredentials([ file(credentialsId: "${IMG_PULL_SECRET_CRED}", variable: 'PULL_SECRET_FILE'),
                      file(credentialsId: "${K8S_CRED}", variable: 'SECRET_FILE')]) {
        sh "kubectl --kubeconfig=${SECRET_FILE} -n ${namespace} apply -f ${PULL_SECRET_FILE} "
    }
}

// Get service name from helm template
def getServiceName(tpl){
    if (tpl.getClass() == java.util.LinkedList){
        result = ""
        tpl.each{ item ->
            def port_check = item.spec.ports.find{it.name == "http"}
            if (port_check){
                result = "http://${item.metadata.name}:${port_check.port}"
            }
        }
        return result
    }
    else {
        def port_check = tpl.spec.ports.find{it.name == "http"}
        return "http://${tpl.metadata.name}:${port_check.port}"
    }
}
