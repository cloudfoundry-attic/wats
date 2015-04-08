set -e

cd `dirname $0`

export GOPATH=$PWD/..
export GOBIN=$GOPATH/bin

go get github.com/onsi/ginkgo/ginkgo
# The following line will fail with the || echo, since tests don't
# have a binary and go get will try to build one
go get -t ../tests/... 2>/dev/null || echo "Installed dependencies"
ginkgo -r -failFast -slowSpecThreshold=120 $@ ../tests
