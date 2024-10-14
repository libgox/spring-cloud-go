package springcloud

type Endpoint struct {
	Name                string `json:"name"`
	Id                  string `json:"id"`
	Address             string `json:"address"`
	Port                int    `json:"port"`
	SslPort             *int   `json:"sslPort"`
	RegistrationTimeUTC int64  `json:"registrationTimeUTC"`
	ServiceType         string `json:"serviceType"`
}
