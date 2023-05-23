package download

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestVersion(t *testing.T) {
	_, err := LatestVersion("")
	if err == nil {
		t.Fatalf(`LatestVersion("") err: %v`, err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", `<!DOCTYPE html>
<head>
  <title>Download LibreOffice | LibreOffice - Free Office Suite - Based on OpenOffice - Compatible with Microsoft</title>
</head>
<body class="Download" id="download-libreoffice">
  <span class="dl_version_number">7.5.3</span><br />
  <span class="dl_description_text">If you're a technology enthusiast, early adopter or power user, this version is for you!</span>
  <span class="dl_version_number">7.4.7</span><br />
  <span class="dl_description_text">This version is slightly older and does not have the latest features, but it has been tested for longer. For business deployments, we <a href="https://www.libreoffice.org/download/libreoffice-in-business/">strongly recommend support from certified partners</a> which also offer long-term support versions of LibreOffice.</span>
</body>
</html>
`)
	}))
	defer server.Close()
	want := "7.4.7"
	got, err := LatestVersion(server.URL)
	if err != nil {
		t.Fatalf("LatestVersion() err: %v", err)
	}
	if got != want {
		t.Fatalf("LatestVersion() = %q, want: %q", got, want)
	}
}
