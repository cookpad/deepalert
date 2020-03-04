package inspector

var (
	HandleTask           = handleTask
	ExtractRegionFromURL = extractRegionFromURL
)

func InjectNewSQSClient(client sqsClient) {
	newSqsClient = func(region string) sqsClient {
		return client
	}
}

func FixNewSQSClient() {
	newSqsClient = newAwsSqsClient
}
