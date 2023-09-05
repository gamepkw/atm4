$repositoryUrl = "https://github.com/gamepkw/atm4.git"
$destinationDirectory = "C:/Users/admin/Project_GOs/atm4_deploy3"
git clone $repositoryUrl $destinationDirectory
cd $destinationDirectory
npm install
npm run build
npm test
npm start
