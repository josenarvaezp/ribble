pipeline {
    agent any
    stages {
        stage('Start localstack') {
            steps {
                // echo "YES"
                start localstack
                timeout(time: 1, unit: 'MINUTES') {
                    sh './build/integration_tests/localstack.sh'
                }

                // create integration test bucket
                sh 'make integration-s3'
            }
        }

        stage('Run test 1') {
            steps {
                sh 'go test -run TestBuildQ1 ./build/integration_tests/fts'
                sh 'go test -run TestUploadQ1 ./build/integration_tests/fts'
                timeout(time: 3, unit: 'MINUTES') {
                    sh 'go test -run TestRunQ1 ./build/integration_tests/fts'
                }
            }
        }
    }
    // post {
    //     cleanup {
    //         sh 'docker-compose down'
    //         cleanWs()
    //     }
    // }
}

// pipeline {
//     agent none
//     stages {
//         stage('Start localstack') {
//             agent any
//             steps {
//                 sh 'echo "Hello World"'

//                 // start localstack
//                 timeout(time: 1, unit: 'MINUTES') {
//                     sh './build/integration_tests/localstack.sh'
//                 }

//                 // create integration test bucket
//                 sh 'make integration-s3'
//             }
//         }

//         stage('Build ribble') {
//             agent any
//             steps {
//                 sh 'make build_cli'
//             }
//         }

//         stage('Run test 1') {
//             agent any
//             steps {
//                 sh './build/integration_tests/test1.sh'
//             }
//         }


//     }
// }