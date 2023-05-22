rm ../../../utils/mc
rm ../../../utils/cp
rm ../../../utils/sp
rm ../../../utils/ms

go build -o mc mc/mc.go
go build -o cp cp/cp.go
go build -o sp sp/sp.go
go build -o ms ms/ms.go

mv mc/mc ../../../utils/mc
mv cp/cp ../../../utils/cp
mv sp/sp ../../../utils/sp
mv ms/ms ../../../utils/ms