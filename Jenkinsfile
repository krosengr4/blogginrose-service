pipeline {
    agent any

    environment {
        DOCKER_IMAGE = 'krosengr4/blogginrose-service'
        DOCKER_CREDENTIALS_ID = 'dockerhub-creds'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    echo 'Building Docker Image...'
                    sh "docker build -t ${DOCKER_IMAGE}:${BUILD_NUMBER} -t ${DOCKER_IMAGE}:latest ."
                }
            }
        }

        stage ('Push to Docker Hub') {
            steps {
                script {
                    echo 'Pushing to Docker Hub...'
                    withCredentials([usernamePassword(
                        credentialsId: DOCKER_CREDENTIALS_ID,
                        usernameVariable: 'DOCKER_USER',
                        passwordVariable: 'DOCKER_PASS'
                    )]) {
                        sh '''
                            echo $DOCKER_PASS | docker login -u $DOCKER_USER --password-stdin
                            docker push ${DOCKER_IMAGE}:${BUILD_NUMBER}
                            docker push ${DOCKER_IMAGE}:latest
                        '''
                    }
                }
            }
        }
    }

    post {
        success {
            echo "Build successful! Image: ${DOCKER_IMAGE}:${BUILD_NUMBER}"
            build job: 'deploy', wait: false
        } 
        failure {
            echo '!!! BUILD FAILED !!!'
        } 
        always {
            sh 'docker logout'
        }
    }
}