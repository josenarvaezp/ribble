pipeline {
    agent none
    stages {
        stage('Start localstack') {
            agent any
            steps {
                sh 'echo "Hello World"'

                // start localstack
                timeout(time: 1, unit: 'MINUTES') {
                    sh './build/integration-tests/localstack.sh'
                }

                // create integration test bucket
                sh 'make integration-s3'
            }
        }
        // stage('Test') {
        //     agent {
        //         docker { image 'integration:latest' }
        //     }
        //     steps {
        //         sh 'echo "HERE"'
        //         sh 'make test'
        //     }
        // }
    }
}

// pipeline {
//     agent none
//     stages {
//         stage('Build docker') {
//             agent any
//             steps {
//                 sh 'echo "Hello World"'
//                 sh '''
//                     echo "Multiline shell steps works too"
//                     awslocal s3 ls
//                 '''
//                 sh 'make build-integration'
//             }
//         }
//         stage('Test') {
//             agent {
//                 docker { image 'integration:latest' }
//             }
//             steps {
//                 sh 'echo "HERE"'
//                 sh 'make test'
//             }
//         }
//     }
// }