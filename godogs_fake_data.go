package godog

//random value indices that value of given key should be generated randomly
const random = "random"

//FakeUsername represents fake user username
type FakeUsername string

//FakePassword represents fake user password
type FakePassword string

//Data holds user representation of data for given suite
type Data struct {
	Users []FakeUser `json:"users"`
}

//Generator implements method for generating random data
type Generator struct{}

//FakeUser represents fake user
type FakeUser struct {
	//Alias indices variable name under which data will be saved
	Alias    string       `json:"alias"`
	Username FakeUsername `json:"username"`
	Password FakePassword `json:"password"`
}

//RandomUsername generates random username
func (u Generator) RandomUsername(af *ApiFeature) FakeUsername {
	return FakeUsername(af.stringWithCharset(10, charset))
}

//RandomPassword generates random password
func (u Generator) RandomPassword(af *ApiFeature) FakePassword {
	return FakePassword(af.stringWithCharset(10, charset))
}

//shouldGenerate indices whether username should be generated or not
func (f FakeUsername) shouldGenerate() bool {
	return len(f) == 0 || string(f) == random
}

//shouldGenerate indices whether password should be generated or not
func (f FakePassword) shouldGenerate() bool {
	return len(f) == 0 || string(f) == random
}

//Generate is responsible for fulfilling all fields that suppose to be randomly generated
func (d *Data) Generate(af *ApiFeature) {
	generator := Generator{}

	for i, user := range d.Users {
		if user.Username.shouldGenerate() {
			d.Users[i].Username = generator.RandomUsername(af)
		}

		if user.Password.shouldGenerate() {
			d.Users[i].Password = generator.RandomPassword(af)
		}
	}
}
