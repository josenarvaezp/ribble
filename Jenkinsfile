pipeline {
    agent any
    stages {
        stage('Start localstack') {
            steps {
                // start localstack
                timeout(time: 1, unit: 'MINUTES') {
                    sh './build/integration_tests/localstack.sh'
                }

                // create integration test bucket
                sh 'make integration-s3'
            }
        }

        // stage('Run test 1') {
        //     steps {
        //         sh 'go test -run TestBuildQ1 ./build/integration_tests/tests'
        //         sh 'go test -run TestUploadQ1 ./build/integration_tests/tests'
        //         timeout(time: 3, unit: 'MINUTES') {
        //             sh 'go test -run TestRunQ1 ./build/integration_tests/tests'
        //         }
        //     }
        // }

        stage('Run test 2') {
            steps {
                sh 'go test -run TestBuildQ6 ./build/integration_tests/tests'
                sh 'go test -run TestUploadQ6 ./build/integration_tests/tests'
                timeout(time: 3, unit: 'MINUTES') {
                    sh 'go test -run TestRunQ6 ./build/integration_tests/tests'
                }
            }
        }
    }
    post {
        cleanup {
            sh 'docker-compose -f ./build/integration_tests/docker-compose.yml down'
            cleanWs()
        }
    }
}