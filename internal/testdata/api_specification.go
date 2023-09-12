package testdata

import (
	"fmt"
	"strings"

	"github.com/liveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

var (
	APISpecificationResponseBody  = ToJSON(APISpecifications[0])
	APISpecificationsResponseBody = ToJSON(APISpecifications[0])
	APISpecificationsNoCategory   = APISpecifications[2]

	APISpecificationDefinition = `{
			"openapi": "3.0.0",
			"info": {
				"version": "1.1.1",
				"title": "Test API Spec",
				"license": {
					"name": "MIT"
				}
			}
		}
		`

	APISpecificationDefinitionSrc = func() string {
		escapeQuotes := strings.ReplaceAll(APISpecificationDefinition, `"`, `\"`)
		escapedNewlines := strings.ReplaceAll(escapeQuotes, "\n", "\\n")

		return escapedNewlines
	}()

	APIRegistryResponseBodyCreated = `{"registryUUID": "abcdefghijklmno", "definition": ` + APISpecificationDefinition + `}`

	APISpecificationSavedResponse = fmt.Sprintf(
		"{\"_id\":\"%s\",\"title\":\"%s\"}",
		APISpecifications[0].ID,
		APISpecifications[0].Title,
	)
)

var APISpecifications = []readme.APISpecification{
	{
		ID:         "6398a4a594b26e00885e7ec0",
		LastSynced: "2022-12-13T16:41:39.512Z",
		Category: readme.CategorySummary{
			ID:    "63f8dc63d70452003b73ff12",
			Title: "Test API Spec",
			Slug:  "test-api-spec",
		},
		Source:  "api",
		Title:   "Test API Spec",
		Type:    "oas",
		Version: "638cf4cfdea3ff0096d1a95a",
	},
	{
		ID:         "6398a4a594b26e00885e7ec1",
		LastSynced: "2022-12-13T16:40:39.512Z",
		Category: readme.CategorySummary{
			ID:    "63f8dc63d70452003b73ff13",
			Title: "Another Test API Spec",
			Slug:  "another-test-api-spec",
		},
		Source:  "api",
		Title:   "Another Test API Spec",
		Type:    "oas",
		Version: "638cf4cfdea3ff0096d1a95b",
	},
	{
		ID:         "6398a4a594b26e00885e7ec2",
		LastSynced: "2022-12-13T16:39:39.512Z",
		Source:     "api",
		Title:      "Test API Spec Without Category",
		Type:       "oas",
		Version:    "638cf4cfdea3ff0096d1a95c",
	},
}

func APISpecificationRespond(response any, code int) func() {
	return func() {
		gock.OffAll()
		gock.New(testURL).
			Get("/api-specification").
			MatchParam("perPage", "100").
			MatchParam("page", "1").
			Persist().
			Reply(code).
			SetHeaders(map[string]string{"link": `<>; rel="next", <>; rel="prev", <>; rel="last"`}).
			JSON(response)
	}
}

func APISpecificationCreateRespond(mockVersionList []readme.VersionSummary) func() {
	return func() {
		gock.OffAll()
		// Create in the registry.
		gock.New(testURL).Post("/api-registry").Times(1).Reply(201).JSON(APIRegistryResponseBodyCreated)
		gock.New(testURL).Get("/api-registry").Times(1).Reply(200).JSON(APISpecificationDefinition)

		// Create the API spec.
		gock.New(testURL).Post("/api-specification").Times(1).Reply(201).JSON(APISpecificationSavedResponse)

		// Lookup version
		gock.New(testURL).Get("/version").Times(1).Reply(200).JSON(mockVersionList)
		gock.New(testURL).
			Get("/version" + "/" + mockVersionList[0].VersionClean).
			Times(1).
			Reply(200).
			JSON(mockVersionList[0])

		// Get the created API spec.
		gock.New(testURL).
			Get("/api-specification").
			MatchParam("perPage", "100").
			MatchParam("page", "1").
			Persist().
			Reply(200).
			SetHeaders(map[string]string{"link": `<>; rel="next", <>; rel="prev", <>; rel="last"`}).
			JSON(APISpecifications)

		// Delete the API spec.
		gock.New(testURL).Delete("/api-specification").Times(1).Reply(204)
	}
}

func APISpecificationDeleteCategoryRespond(mockVersionList []readme.VersionSummary) func() {
	return func() {
		APISpecificationCreateRespond(mockVersionList)()

		// Delete the category.
		gock.New(testURL).Delete("/categories/" + APISpecifications[0].Category.Slug).Times(1).Reply(204)
	}
}
