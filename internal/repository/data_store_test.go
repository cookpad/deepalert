package repository_test

/*
func TestDataStoreTakeReportID(t *testing.T) {
	cfg := test.LoadTestConfig()
	ts := time.Now().UTC()
	detector := uuid.New().String()

	alert1 := da.Alert{
		Detector:  detector,
		RuleName:  "myRule",
		RuleID:    "test1",
		AlertKey:  "blue",
		Timestamp: ts,
	}
	alert2 := da.Alert{
		Detector:  detector,
		RuleName:  "myRule",
		RuleID:    "test1",
		AlertKey:  "blue",
		Timestamp: ts.Add(time.Hour * 1),
	}
	alert3 := da.Alert{
		Detector:  detector,
		RuleName:  "myRule",
		RuleID:    "test1",
		AlertKey:  "orange",
		Timestamp: ts.Add(time.Hour * 4),
	}

	svc := f.NewDataStoreService(cfg.TableName, cfg.Region)

	report1, err := svc.TakeReport(alert1)
	require.NoError(t, err)
	assert.NotNil(t, report1)
	assert.True(t, report1.IsNew())

	report2, err := svc.TakeReport(alert2)
	assert.NotNil(t, report2)
	require.NoError(t, err)
	assert.False(t, report2.IsNew())
	assert.True(t, report2.IsMore())

	// Another result of 1 hour later with same alertID should have same ReportID
	assert.Equal(t, report1.ID, report2.ID)

	report3, err := svc.TakeReport(alert3)
	assert.NotNil(t, report3)
	require.NoError(t, err)
	assert.True(t, report3.IsNew())
	// However result over 3 hour later with same alertID should have other ReportID
	assert.NotEqual(t, report1.ID, report3.ID)
}

func TestDataStoreAlertCache(t *testing.T) {
	cfg := test.LoadTestConfig()
	svc := f.NewDataStoreService(cfg.TableName, cfg.Region)

	alert1 := da.Alert{
		Detector:  "me",
		RuleName:  "myRule",
		RuleID:    "test1",
		AlertKey:  "blue",
		Timestamp: time.Now(),
	}
	alert2 := da.Alert{
		Detector:  "you",
		RuleName:  "myRule",
		RuleID:    "test2",
		AlertKey:  "orange",
		Timestamp: time.Now(),
	}
	alert3 := da.Alert{
		Detector:  "someone",
		RuleName:  "myRule",
		RuleID:    "test3",
		AlertKey:  "gray",
		Timestamp: time.Now(),
	}

	var err error
	reportID := f.NewReportID()
	err = svc.SaveAlertCache(reportID, alert1)
	require.NoError(t, err)
	err = svc.SaveAlertCache(reportID, alert2)
	require.NoError(t, err)

	anotherReportID := f.NewReportID()
	err = svc.SaveAlertCache(anotherReportID, alert3)
	require.NoError(t, err)

	alerts, err := svc.FetchAlertCache(reportID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(alerts))

	assert.True(t, alerts[0].Detector == "me" || alerts[1].Detector == "me")
	assert.True(t, alerts[0].Detector == "you" || alerts[1].Detector == "you")
}

func TestDataStoreReportContent(t *testing.T) {
	cfg := test.LoadTestConfig()
	svc := f.NewDataStoreService(cfg.TableName, cfg.Region)

	rID1 := f.NewReportID()
	rID2 := f.NewReportID()

	section1 := da.ReportSection{
		ReportID: rID1,
		Author:   "blue",
		Attribute: da.Attribute{
			Type:  da.TypeIPAddr,
			Key:   "Remote host",
			Value: "10.1.2.3",
		},
	}
	section2 := da.ReportSection{
		ReportID: rID1,
		Author:   "orange",
		Attribute: da.Attribute{
			Type:  da.TypeIPAddr,
			Key:   "Remote host",
			Value: "10.1.2.3",
		},
	}
	section3 := da.ReportSection{
		ReportID: rID2,
		Author:   "orange",
		Attribute: da.Attribute{
			Type:  da.TypeIPAddr,
			Key:   "Remote host",
			Value: "10.1.2.3",
		},
	}

	err := svc.SaveReportSection(section1)
	require.NoError(t, err)
	err = svc.SaveReportSection(section2)
	require.NoError(t, err)
	err = svc.SaveReportSection(section3)
	require.NoError(t, err)

	sections, err := svc.FetchReportSection(rID1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(sections))
	assert.Equal(t, rID1, sections[0].ReportID)
	assert.Equal(t, rID1, sections[1].ReportID)
}
*/
