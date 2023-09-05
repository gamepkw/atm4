# $repositoryUrl = "https://github.com/gamepkw/atm4.git"
# $destinationDirectory = "C:/Users/admin/Project_GOs/atm4_deploy3"
# git clone $repositoryUrl $destinationDirectory
# cd $destinationDirectory
# npm install
# npm run build
# npm test
# npm start

pipeline {
    agent any // You can specify the agent to run on

    stages {
        stage('Checkout') {
            steps {
                checkout scm // Checkout the source code
            }
        }

        stage('Set up Node.js') {
            steps {
                // Configure Node.js setup (if required)
            }
        }

        stage('Install Dependencies') {
            steps {
                sh 'npm install' // Run npm install
            }
        }

        stage('Build Project') {
            steps {
                sh 'npm run build' // Run npm run build
            }
        }

        stage('Run Tests') {
            steps {
                sh 'npm test' // Run npm test
            }
        }

        stage('Deploy to Production') {
            steps {
                sh './deploy.ps1' // Run the PowerShell script
            }
        }
    }
}

