package load_balance

type LoadBalance interface{
	Add(addr... string) error
	Get(clientAddr string)string
}

type Observer interface{
	Update()
}