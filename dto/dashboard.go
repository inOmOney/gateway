package dto

type ServiceStatOutput struct{
	Today []int64	`json:"today"`
	Yesterday []int64 `json:"yesterday"`
}

type DashServiceStatOutput struct{
	Data []DashServiceStatItemOutput`json:"data"`
	Legend []string`json:"legend"`
}

type DashServiceStatItemOutput struct {
	LoadType int `json:"load_type"`
	Name string`json:"name"`
	Value int`json:"value"`
}

type DashPanelOutput struct{
	CurrentQps int64	`json:"currentQps"`
	TodayRequestNum int64 `json:"todayRequestNum"`
	ServiceNum int `json:"serviceNum"`
	AppNum int `json:"appNum"`
}