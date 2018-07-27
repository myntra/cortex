echo "yarn and rice"
rm -rf cmd/build
cd ui && yarn install && yarn run build && cd ..
cd cmd && rice embed-go && cd ..