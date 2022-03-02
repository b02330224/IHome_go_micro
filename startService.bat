start cmd /k  go run  -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" .\IhomeWeb\main.go --registry=consul

start cmd /k  go run .\GetArea\main.go --registry=consul
start cmd /k  go run .\GetImageCd\main.go --registry=consul
start cmd /k  go run .\PostRet\main.go --registry=consul
start cmd /k  go run .\PostLogin\main.go --registry=consul
start cmd /k  go run .\PostUserAuth\main.go  --registry=consul
start cmd /k  go run .\GetSession\main.go --registry=consul
start cmd /k  go run .\DeleteSession\main.go --registry=consul
start cmd /k  go run .\PostAvatar\main.go  --registry=consul
start cmd /k  go run .\GetUserInfo\main.go --registry=consul
start cmd /k  go run .\PostUserAuth\main.go --registry=consul
start cmd /k  go run .\PostAvatar\main.go --registry=consul
start cmd /k  go run .\GetUserHouses\main.go --registry=consul
start cmd /k  go run .\PostHouses\main.go --registry=consul
start cmd /k  go run .\PostHousesImage\main.go --registry=consul
start cmd /k  go run .\GetIndex\main.go --registry=consul
start cmd /k  go run .\GetHouses\main.go --registry=consul

start cmd /k  go run .\PutComment\main.go --registry=consul
start cmd /k  go run .\PostOrders\main.go --registry=consul

start cmd /k  go run .\GetUserOrder\main.go --registry=consul

start cmd /k  go run .\PutOrders\main.go --registry=consul