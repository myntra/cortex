echo "yarn and rice"
cd ui && yarn install && yarn run build && cd ..
cd cmd && rice embed-go && cd ..