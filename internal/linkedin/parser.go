package linkedin

import (
	"fmt"
	"strings"

	"github.com/Decodx09/linkedin-profile-api/internal/models"
)

func str(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func getMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	return asMap(m[key])
}

func getSlice(m map[string]any, key string) []any {
	if m == nil {
		return nil
	}
	if s, ok := m[key].([]any); ok {
		return s
	}
	return nil
}

func elements(profileView map[string]any, section string) []any {
	return getSlice(getMap(profileView, section), "elements")
}

func numToInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	}
	return 0, false
}

func fmtDate(node map[string]any) string {
	if node == nil {
		return ""
	}
	year, hasYear := numToInt(node["year"])
	month, hasMonth := numToInt(node["month"])
	switch {
	case hasYear && hasMonth:
		return fmt.Sprintf("%04d-%02d", year, month)
	case hasYear:
		return fmt.Sprintf("%04d", year)
	}
	return ""
}

func dateRange(timePeriod map[string]any) models.DateRange {
	return models.DateRange{
		Start: fmtDate(getMap(timePeriod, "startDate")),
		End:   fmtDate(getMap(timePeriod, "endDate")),
	}
}

func bestImageURL(node map[string]any) string {
	if node == nil {
		return ""
	}
	if vi := getMap(node, "com.linkedin.common.VectorImage"); vi != nil {
		node = vi
	}
	root := str(node, "rootUrl")
	artifacts := getSlice(node, "artifacts")
	if root == "" || len(artifacts) == 0 {
		return ""
	}
	bestWidth := -1
	bestSeg := ""
	for _, a := range artifacts {
		am := asMap(a)
		if am == nil {
			continue
		}
		w, _ := numToInt(am["width"])
		if w > bestWidth {
			bestWidth = w
			bestSeg = str(am, "fileIdentifyingUrlPathSegment")
		}
	}
	if bestSeg == "" {
		return root
	}
	return root + bestSeg
}

func parseExperience(els []any) []models.Experience {
	out := make([]models.Experience, 0, len(els))
	for _, e := range els {
		el := asMap(e)
		if el == nil {
			continue
		}
		company := getMap(el, "company")
		var companyURL string
		if urn := str(el, "companyUrn"); urn != "" {
			parts := strings.Split(urn, ":")
			id := parts[len(parts)-1]
			if id != "" {
				companyURL = "https://www.linkedin.com/company/" + id + "/"
			}
		}
		out = append(out, models.Experience{
			Title:              str(el, "title"),
			Company:            str(el, "companyName"),
			CompanyLinkedInURL: companyURL,
			EmploymentType:     str(el, "employmentType"),
			Location:           firstNonEmpty(str(el, "locationName"), str(el, "geoLocationName")),
			Description:        str(el, "description"),
			DateRange:          dateRange(getMap(el, "timePeriod")),
			CompanyLogo:        bestImageURL(getMap(getMap(company, "miniCompany"), "logo")),
		})
	}
	return out
}

func parseEducation(els []any) []models.Education {
	out := make([]models.Education, 0, len(els))
	for _, e := range els {
		el := asMap(e)
		if el == nil {
			continue
		}
		out = append(out, models.Education{
			School:       str(el, "schoolName"),
			Degree:       str(el, "degreeName"),
			FieldOfStudy: str(el, "fieldOfStudy"),
			Description:  str(el, "description"),
			Grade:        str(el, "grade"),
			Activities:   str(el, "activities"),
			DateRange:    dateRange(getMap(el, "timePeriod")),
			SchoolLogo:   bestImageURL(getMap(getMap(el, "school"), "logo")),
		})
	}
	return out
}

func parseCertifications(els []any) []models.Certification {
	out := make([]models.Certification, 0, len(els))
	for _, e := range els {
		el := asMap(e)
		if el == nil {
			continue
		}
		out = append(out, models.Certification{
			Name:          str(el, "name"),
			Authority:     str(el, "authority"),
			LicenseNumber: str(el, "licenseNumber"),
			URL:           str(el, "url"),
			DateRange:     dateRange(getMap(el, "timePeriod")),
		})
	}
	return out
}

func parseLanguages(els []any) []models.Language {
	out := make([]models.Language, 0, len(els))
	for _, e := range els {
		el := asMap(e)
		if el == nil {
			continue
		}
		out = append(out, models.Language{
			Name:        str(el, "name"),
			Proficiency: str(el, "proficiency"),
		})
	}
	return out
}

func parseVolunteer(els []any) []models.Volunteer {
	out := make([]models.Volunteer, 0, len(els))
	for _, e := range els {
		el := asMap(e)
		if el == nil {
			continue
		}
		out = append(out, models.Volunteer{
			Role:         str(el, "role"),
			Organization: str(el, "companyName"),
			Cause:        str(el, "cause"),
			Description:  str(el, "description"),
			DateRange:    dateRange(getMap(el, "timePeriod")),
		})
	}
	return out
}

func parsePublications(els []any) []models.Publication {
	out := make([]models.Publication, 0, len(els))
	for _, e := range els {
		el := asMap(e)
		if el == nil {
			continue
		}
		out = append(out, models.Publication{
			Name:        str(el, "name"),
			Publisher:   str(el, "publisher"),
			Description: str(el, "description"),
			URL:         str(el, "url"),
			Date:        fmtDate(getMap(el, "date")),
		})
	}
	return out
}

func parseHonors(els []any) []models.Honor {
	out := make([]models.Honor, 0, len(els))
	for _, e := range els {
		el := asMap(e)
		if el == nil {
			continue
		}
		out = append(out, models.Honor{
			Title:       str(el, "title"),
			Issuer:      str(el, "issuer"),
			Description: str(el, "description"),
			Date:        fmtDate(getMap(el, "issueDate")),
		})
	}
	return out
}

func parseSkillNames(els []any) []string {
	out := make([]string, 0, len(els))
	for _, e := range els {
		if name := str(asMap(e), "name"); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func parseContactInfo(contact map[string]any) models.ContactInfo {
	ci := models.ContactInfo{
		Emails:       []string{},
		PhoneNumbers: []string{},
		Websites:     []string{},
		Twitter:      []string{},
	}
	if contact == nil {
		return ci
	}
	if email := str(contact, "emailAddress"); email != "" {
		ci.Emails = append(ci.Emails, email)
	}
	for _, w := range getSlice(contact, "websites") {
		if u := str(asMap(w), "url"); u != "" {
			ci.Websites = append(ci.Websites, u)
		}
	}
	for _, p := range getSlice(contact, "phoneNumbers") {
		if num := str(asMap(p), "number"); num != "" {
			ci.PhoneNumbers = append(ci.PhoneNumbers, num)
		}
	}
	for _, t := range getSlice(contact, "twitterHandles") {
		if name := str(asMap(t), "name"); name != "" {
			ci.Twitter = append(ci.Twitter, name)
		}
	}
	return ci
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func ParseProfile(publicID string, profileView, contactInfo, skillsPayload map[string]any) models.Profile {
	profileNode := getMap(profileView, "profile")
	mini := getMap(profileNode, "miniProfile")

	first := str(profileNode, "firstName")
	last := str(profileNode, "lastName")
	full := strings.TrimSpace(strings.Join(nonEmpty(first, last), " "))

	skills := parseSkillNames(getSlice(skillsPayload, "elements"))
	if len(skills) == 0 {
		skills = parseSkillNames(elements(profileView, "skillView"))
	}

	profilePic := bestImageURL(getMap(mini, "picture"))
	if profilePic == "" {
		profilePic = bestImageURL(getMap(profileNode, "picture"))
	}

	return models.Profile{
		PublicID:        publicID,
		ProfileURL:      "https://www.linkedin.com/in/" + publicID + "/",
		FirstName:       first,
		LastName:        last,
		FullName:        full,
		Headline:        str(profileNode, "headline"),
		Summary:         str(profileNode, "summary"),
		Location:        firstNonEmpty(str(profileNode, "locationName"), str(profileNode, "geoLocationName")),
		Country:         str(profileNode, "geoCountryName"),
		Industry:        str(profileNode, "industryName"),
		ProfilePicture:  profilePic,
		BackgroundImage: bestImageURL(getMap(mini, "backgroundImage")),
		Experience:      parseExperience(elements(profileView, "positionView")),
		Education:       parseEducation(elements(profileView, "educationView")),
		Skills:          skills,
		Certifications:  parseCertifications(elements(profileView, "certificationView")),
		Languages:       parseLanguages(elements(profileView, "languageView")),
		Volunteer:       parseVolunteer(elements(profileView, "volunteerExperienceView")),
		Publications:    parsePublications(elements(profileView, "publicationView")),
		Honors:          parseHonors(elements(profileView, "honorView")),
		ContactInfo:     parseContactInfo(contactInfo),
	}
}

func nonEmpty(vals ...string) []string {
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
