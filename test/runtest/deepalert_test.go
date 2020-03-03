package deepalert_test

/*
func TestNormalWorkFlow(t *testing.T) {
	t.Skip()

	cfg := test.LoadTestConfig()
	alertKey := uuid.New().String()

	alert := deepalert.Alert{
		Detector:  "test",
		RuleName:  "TestRule",
		RuleID:    "xxx",
		AlertKey:  alertKey,
		Timestamp: time.Now().UTC(),
		Attributes: []deepalert.Attribute{
			{
				Type:    deepalert.TypeIPAddr,
				Key:     "test value",
				Value:   "192.168.0.1",
				Context: []deepalert.AttrContext{deepalert.CtxLocal},
			},
		},
	}
	alertMsg, err := json.Marshal(alert)
	require.NoError(t, err)

	var reportID string
	gp.SetLoggerTraceLevel()

	playbook := []gp.Scene{
		// Send request
		gp.PublishSnsMessage(gp.LogicalID("AlertNotification"), alertMsg),
		gp.GetLambdaLogs(gp.LogicalID("ReceptAlert"), func(log gp.CloudWatchLog) bool {
			assert.Contains(t, log, alertKey)
			return true
		}).Filter(alertKey),
		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			var entry struct {
				ReportID string `dynamo:"report_id"`
			}

			alertID := "alertmap/" + alert.AlertID()
			err := table.Get("pk", alertID).Range("sk", dynamo.Equal, "Fixed").One(&entry)
			if err != nil {
				return false
			}
			require.NotEmpty(t, entry.ReportID)
			reportID = entry.ReportID
			fmt.Println("reportID:", reportID)
			return true
		}),
		gp.GetLambdaLogs(gp.LogicalID("DispatchInspection"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
		gp.GetLambdaLogs(gp.LogicalID("SubmitContent"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
		gp.GetLambdaLogs(gp.LogicalID("FeedbackAttribute"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
		gp.GetLambdaLogs(gp.LogicalID("FeedbackAttribute"), func(log gp.CloudWatchLog) bool {
			return log.Contains("mizutani")
		}),
		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			var caches []struct {
				Key   string `dynamo:"attr_key"`
				Value string `dynamo:"attr_value"`
				Type  string `dynamo:"attr_type"`
			}

			pk := "attribute/" + reportID
			if err := table.Get("pk", pk).All(&caches); err != nil {
				return false
			}

			if len(caches) != 2 {
				return false
			}

			var a1, a2 int
			if caches[0].Type == "ipaddr" {
				a1, a2 = 0, 1
			} else {
				a1, a2 = 1, 0
			}

			assert.Equal(t, "192.168.0.1", caches[a1].Value)
			assert.Equal(t, "mizutani", caches[a2].Value)
			assert.Equal(t, "username", caches[a2].Type)
			return true
		}),

		gp.GetDynamoRecord(gp.LogicalID("CacheTable"), func(table dynamo.Table) bool {
			var contents []struct {
				Data []byte `dynamo:"data"`
			}

			pk := "content/" + reportID

			if err := table.Get("pk", pk).All(&contents); err != nil {
				return false
			}

			require.True(t, len(contents) > 0)
			require.NotEmpty(t, contents[0].Data)
			return true
		}),

		gp.Pause(10),

		gp.GetLambdaLogs(gp.LogicalID("CompileReport"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
		gp.GetLambdaLogs(gp.LogicalID("PublishReport"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
	}

	err = gp.New(cfg.Region, cfg.StackName).Play(playbook)
	require.NoError(t, err)
}

func TestNormalAggregation(t *testing.T) {
	t.Skip()
	cfg := test.LoadTestConfig()
	alertKey := uuid.New().String()
	attr1 := uuid.New().String()
	attr2 := uuid.New().String()

	alert := deepalert.Alert{
		Detector:  "test",
		RuleName:  "TestRule",
		RuleID:    "yyy",
		AlertKey:  alertKey,
		Timestamp: time.Now().UTC(),
		Attributes: []deepalert.Attribute{
			{
				Type:    deepalert.TypeUserName,
				Key:     "blue",
				Value:   attr1,
				Context: []deepalert.AttrContext{deepalert.CtxLocal},
			},
		},
	}
	alertMsg1, err := json.Marshal(alert)
	alert.Attributes = []deepalert.Attribute{
		{
			Type:    deepalert.TypeUserName,
			Key:     "orange",
			Value:   attr2,
			Context: []deepalert.AttrContext{deepalert.CtxLocal},
		},
	}
	alertMsg2, err := json.Marshal(alert)

	require.NoError(t, err)

	var reportID string

	playbook := []gp.Scene{
		// Send request
		gp.PublishSnsMessage(gp.LogicalID("AlertNotification"), alertMsg1),
		gp.PublishSnsMessage(gp.LogicalID("AlertNotification"), alertMsg2),
		gp.GetLambdaLogs(gp.LogicalID("ReceptAlert"), func(log gp.CloudWatchLog) bool {
			assert.Contains(t, log, alertKey)
			return true
		}).Filter(alertKey),

		gp.Pause(10),

		gp.GetLambdaLogs(gp.LogicalID("CompileReport"), func(log gp.CloudWatchLog) bool {
			return log.Contains(reportID)
		}),
		gp.GetLambdaLogs(gp.Arn(cfg.TestPublisherArn), func(log gp.CloudWatchLog) bool {
			return log.Contains(attr1)
		}),
		gp.GetLambdaLogs(gp.Arn(cfg.TestPublisherArn), func(log gp.CloudWatchLog) bool {
			return log.Contains(attr2)
		}),
	}

	err = gp.New(cfg.Region, cfg.StackName).Play(playbook)
	require.NoError(t, err)

}
*/
