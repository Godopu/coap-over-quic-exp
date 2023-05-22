rm server.txt
rm client.txt
go run main.go -exptype tcp -proxytype sp -Nd 30 -bind :8282 >> server.txt &
sleep 2
go run main.go -exptype tcp -proxytype cp -Nd 30 -expTime 5 -spAdr localhost:8282 >> client.txt