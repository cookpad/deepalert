package functions_test

/*
func TestLoggerHook(t *testing.T) {
	cfg := test.LoadTestConfig("..")
	testID1 := uuid.New().String()

	reportID := f.NewReportID()
	now := time.Now().UTC()

	old := f.GetLogOutput()
	f.SetLogOutput(new(bytes.Buffer))
	defer func() { f.SetLogOutput(old) }()

	f.SetLogDestination(cfg.LogGroup, cfg.LogStream, cfg.Region)
	f.SetLoggerContext(nil, "Test", reportID)
	f.Logger.WithField("test_id", testID1).Info("Test")

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
	}))
	cwlogs := cloudwatchlogs.New(ssn)

	detected := func() bool {
		for i := 0; i < 10; i++ {
			resp, err := cwlogs.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(cfg.LogGroup),
				LogStreamName: aws.String(cfg.LogStream),
				StartTime:     aws.Int64(now.Unix() * 1000),
			})

			require.NoError(t, err)
			for _, event := range resp.Events {
				if strings.Index(event.GoString(), testID1) >= 0 {
					return true
				}
			}
			time.Sleep(time.Second * 2)
		}
		return false
	}()

	assert.True(t, detected)
}
*/
