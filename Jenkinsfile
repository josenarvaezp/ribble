pipeline {
    agent none
    stages {
        stage('Build docker') {
            agent any
            steps {
                sh 'echo "Hello World"'
                sh '''
                    echo "Multiline shell steps works too"
                    ls -lah
                '''
                sh 'make build-integration'
            }
        }
        stage('Test') {
            agent {
                docker { image 'integration:latest' }
            }
            steps {
                sh 'echo "HERE"'
                sh 'make test'
            }
        }
    }
}