ALL: transcode remuxing rtmpClient

transcode:transcode.go
	go build transcode.go

remuxing:remuxing.go
	go build remuxing.go

rtmpClient:rtmpClient.go
	go build rtmpClient.go

clean:
	rm transcode
	rm remuxing
	rm rtmpClient
