package linkedin

import (
	"encoding/json"
	"testing"
)

func TestExtractPublicID(t *testing.T) {
	cases := map[string]string{
		"https://www.linkedin.com/in/williamhgates/":         "williamhgates",
		"linkedin.com/in/john-doe-123?originalSubdomain=uk":  "john-doe-123",
		"https://www.linkedin.com/in/satya.nadella/details/": "satya.nadella",
		"satyanadella": "satyanadella",
	}
	for in, want := range cases {
		got, err := ExtractPublicID(in)
		if err != nil {
			t.Fatalf("ExtractPublicID(%q) error: %v", in, err)
		}
		if got != want {
			t.Errorf("ExtractPublicID(%q) = %q, want %q", in, got, want)
		}
	}

	if _, err := ExtractPublicID("https://www.linkedin.com/company/foo/"); err == nil {
		t.Error("expected error for a non-/in/ URL")
	}
}

func TestParseProfile(t *testing.T) {
	const raw = `{
      "profile": {
        "firstName":"Ada","lastName":"Lovelace","headline":"Mathematician",
        "summary":"First programmer.","locationName":"London","geoCountryName":"United Kingdom",
        "industryName":"Software",
        "miniProfile":{"picture":{"com.linkedin.common.VectorImage":{"rootUrl":"https://media.licdn.com/x/",
          "artifacts":[{"width":100,"fileIdentifyingUrlPathSegment":"100.jpg"},
                       {"width":400,"fileIdentifyingUrlPathSegment":"400.jpg"}]}}}
      },
      "positionView":{"elements":[{"title":"Analyst","companyName":"Analytical Engine",
        "locationName":"London","description":"Notes",
        "timePeriod":{"startDate":{"month":5,"year":1842},"endDate":{"year":1843}}}]},
      "educationView":{"elements":[{"schoolName":"Home","degreeName":"Self","fieldOfStudy":"Math"}]},
      "languageView":{"elements":[{"name":"English","proficiency":"NATIVE_OR_BILINGUAL"}]},
      "skillView":{"elements":[{"name":"Algorithms"},{"name":"Mathematics"}]}
    }`

	var pv map[string]any
	if err := json.Unmarshal([]byte(raw), &pv); err != nil {
		t.Fatal(err)
	}
	contact := map[string]any{
		"emailAddress": "ada@x.com",
		"websites":     []any{map[string]any{"url": "https://ada.dev"}},
	}

	p := ParseProfile("adalovelace", pv, contact, nil)

	if p.FullName != "Ada Lovelace" {
		t.Errorf("FullName = %q", p.FullName)
	}
	if p.ProfilePicture != "https://media.licdn.com/x/400.jpg" {
		t.Errorf("ProfilePicture = %q (want largest artifact)", p.ProfilePicture)
	}
	if len(p.Experience) != 1 || p.Experience[0].DateRange.Start != "1842-05" {
		t.Errorf("experience parsed wrong: %+v", p.Experience)
	}
	if p.Experience[0].DateRange.End != "1843" {
		t.Errorf("end date = %q, want 1843", p.Experience[0].DateRange.End)
	}
	if len(p.Skills) != 2 {
		t.Errorf("skills = %v", p.Skills)
	}
	if len(p.ContactInfo.Emails) != 1 || p.ContactInfo.Emails[0] != "ada@x.com" {
		t.Errorf("contact emails = %v", p.ContactInfo.Emails)
	}
	if len(p.ContactInfo.Websites) != 1 {
		t.Errorf("contact websites = %v", p.ContactInfo.Websites)
	}
}
