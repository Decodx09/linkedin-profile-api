package models

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Experience struct {
	Title              string    `json:"title"`
	Company            string    `json:"company"`
	CompanyLinkedInURL string    `json:"company_linkedin_url"`
	EmploymentType     string    `json:"employment_type"`
	Location           string    `json:"location"`
	Description        string    `json:"description"`
	DateRange          DateRange `json:"date_range"`
	CompanyLogo        string    `json:"company_logo"`
}

type Education struct {
	School       string    `json:"school"`
	Degree       string    `json:"degree"`
	FieldOfStudy string    `json:"field_of_study"`
	Description  string    `json:"description"`
	Grade        string    `json:"grade"`
	Activities   string    `json:"activities"`
	DateRange    DateRange `json:"date_range"`
	SchoolLogo   string    `json:"school_logo"`
}

type Certification struct {
	Name          string    `json:"name"`
	Authority     string    `json:"authority"`
	LicenseNumber string    `json:"license_number"`
	URL           string    `json:"url"`
	DateRange     DateRange `json:"date_range"`
}

type Language struct {
	Name        string `json:"name"`
	Proficiency string `json:"proficiency"`
}

type Volunteer struct {
	Role         string    `json:"role"`
	Organization string    `json:"organization"`
	Cause        string    `json:"cause"`
	Description  string    `json:"description"`
	DateRange    DateRange `json:"date_range"`
}

type Publication struct {
	Name        string `json:"name"`
	Publisher   string `json:"publisher"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Date        string `json:"date"`
}

type Honor struct {
	Title       string `json:"title"`
	Issuer      string `json:"issuer"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

type ContactInfo struct {
	Emails       []string `json:"emails"`
	PhoneNumbers []string `json:"phone_numbers"`
	Websites     []string `json:"websites"`
	Twitter      []string `json:"twitter"`
}

type Profile struct {
	PublicID        string `json:"public_id"`
	ProfileURL      string `json:"profile_url"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	FullName        string `json:"full_name"`
	Headline        string `json:"headline"`
	Summary         string `json:"summary"`
	Location        string `json:"location"`
	Country         string `json:"country"`
	Industry        string `json:"industry"`
	ProfilePicture  string `json:"profile_picture"`
	BackgroundImage string `json:"background_image"`

	Experience     []Experience    `json:"experience"`
	Education      []Education     `json:"education"`
	Skills         []string        `json:"skills"`
	Certifications []Certification `json:"certifications"`
	Languages      []Language      `json:"languages"`
	Volunteer      []Volunteer     `json:"volunteer"`
	Publications   []Publication   `json:"publications"`
	Honors         []Honor         `json:"honors"`
	ContactInfo    ContactInfo     `json:"contact_info"`
}

type ProfileResponse struct {
	OK        bool    `json:"ok"`
	Source    string  `json:"source"`
	Cached    bool    `json:"cached"`
	FetchedAt string  `json:"fetched_at"`
	Profile   Profile `json:"profile"`
}
