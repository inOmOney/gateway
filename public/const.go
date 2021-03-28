package public

import "time"

const (
	ValidatorKey    = "ValidatorKey"
	TranslatorKey   = "TranslatorKey"
	AdminSessionKey = "AdminSessionKey"


)

const(
	HttpLoadType   = 0
	UrlRuleType    = 0
	DomainRuleType = 1
	NeedHttps      = 1
	NotNeedHttps   = 0

	TcpLoadType  = 1
	GrpcLoadType = 2

	Deleted = 1
	NoDelete = 0
)

const (
	Interval = 1 * time.Second //流量统计周期
	FlowCountDayServicePrefix = "FlowCountDay_" //Redis中的天数统计 FlowCountDay_28_[ServiceName]
	FlowCountHourServicePrefix = "FlowCountHour_" //Redis中的天数统计 FlowCountHour_28_00_[ServiceName]

	GlobalFlowCount = "GlobalFlowCount"
)
