package bitbucket

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/k0kubun/pp"
	"github.com/mitchellh/mapstructure"
)

type Project struct {
	Key  string
	Name string
}

type Repository struct {
	c *Client

	Project     Project
	Slug        string
	Full_name   string
	Description string
	ForkPolicy  string
	Type        string
	Owner       map[string]interface{}
	Links       map[string]interface{}
}

type RepositoryFile struct {
	Mimetype   string
	Links      map[string]interface{}
	Path       string
	Commit     map[string]interface{}
	Attributes []string
	Type       string
	Size       int
}

type RepositoryBlob struct {
	Content []byte
}

type RepositoryBranches struct {
	Page     int
	Pagelen  int
	MaxDepth int
	Size     int
	Next     string
	Branches []RepositoryBranch
}

type RepositoryBranch struct {
	Type                   string
	Name                   string
	Default_Merge_Strategy string
	Merge_Strategies       []string
	Links                  map[string]interface{}
	Target                 map[string]interface{}
	Heads                  []map[string]interface{}
}

type RepositoryTags struct {
	Page     int
	Pagelen  int
	MaxDepth int
	Size     int
	Next     string
	Tags     []RepositoryTag
}

type RepositoryTag struct {
	Type   string
	Name   string
	Links  map[string]interface{}
	Target map[string]interface{}
	Heads  []map[string]interface{}
}

type Pipeline struct {
	Type       string
	Enabled    bool
	Repository Repository
}

type PipelineVariable struct {
	Type    string
	Uuid    string
	Key     string
	Value   string
	Secured bool
}

type PipelineKeyPair struct {
	Type       string
	Uuid       string
	PublicKey  string
	PrivateKey string
}

type PipelineBuildNumber struct {
	Type string
	Next int
}

type BranchingModel struct {
	Type         string
	Branch_Types []BranchType
	Development  BranchModel
	Production   BranchModel
}

type BranchType struct {
	Kind   string
	Prefix string
}

type BranchModel struct {
	Name           string
	Branch         RepositoryBranch
	Use_Mainbranch bool
}

func (r *Repository) Create(ro *RepositoryOptions) (*Repository, error) {
	data := r.buildRepositoryBody(ro)
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.RepoSlug)
	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeRepository(response)
}

func (r *Repository) Get(ro *RepositoryOptions) (*Repository, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.RepoSlug)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeRepository(response)
}

func (r *Repository) ListFiles(ro *RepositoryFilesOptions) ([]RepositoryFile, error) {
	filePath := path.Join("/repositories", ro.Owner, ro.RepoSlug, "src", ro.Ref, ro.Path) + "/"
	urlStr := r.c.requestUrl(filePath)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeRepositoryFiles(response)
}

func (r *Repository) GetFileBlob(ro *RepositoryBlobOptions) (*RepositoryBlob, error) {
	filePath := path.Join("/repositories", ro.Owner, ro.RepoSlug, "src", ro.Ref, ro.Path)
	urlStr := r.c.requestUrl(filePath)
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}

	blob := RepositoryBlob{Content: content}

	return &blob, nil
}

func (r *Repository) ListBranches(rbo *RepositoryBranchOptions) (*RepositoryBranches, error) {

	params := url.Values{}
	if rbo.Query != "" {
		params.Add("q", rbo.Query)
	}

	if rbo.Sort != "" {
		params.Add("sort", rbo.Sort)
	}

	if rbo.PageNum > 0 {
		params.Add("page", strconv.Itoa(rbo.PageNum))
	}

	if rbo.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(rbo.Pagelen))
	}

	if rbo.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(rbo.MaxDepth))
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/refs/branches?%s", rbo.Owner, rbo.RepoSlug, params.Encode())
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeRepositoryBranches(response)
}

func (r *Repository) ListTags(rbo *RepositoryTagOptions) (*RepositoryTags, error) {

	params := url.Values{}
	if rbo.Query != "" {
		params.Add("q", rbo.Query)
	}

	if rbo.Sort != "" {
		params.Add("sort", rbo.Sort)
	}

	if rbo.PageNum > 0 {
		params.Add("page", strconv.Itoa(rbo.PageNum))
	}

	if rbo.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(rbo.Pagelen))
	}

	if rbo.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(rbo.MaxDepth))
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/refs/tags?%s", rbo.Owner, rbo.RepoSlug, params.Encode())
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeRepositoryTags(response)
}

func (r *Repository) Delete(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.RepoSlug)
	return r.c.execute("DELETE", urlStr, "")
}

func (r *Repository) ListWatchers(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/watchers", ro.Owner, ro.RepoSlug)
	return r.c.execute("GET", urlStr, "")
}

func (r *Repository) ListForks(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/forks", ro.Owner, ro.RepoSlug)
	return r.c.execute("GET", urlStr, "")
}

func (r *Repository) UpdatePipelineConfig(rpo *RepositoryPipelineOptions) (*Pipeline, error) {
	data := r.buildPipelineBody(rpo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config", rpo.Owner, rpo.RepoSlug)
	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineRepository(response)
}

func (r *Repository) AddPipelineVariable(rpvo *RepositoryPipelineVariableOptions) (*PipelineVariable, error) {
	data := r.buildPipelineVariableBody(rpvo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/variables/", rpvo.Owner, rpvo.RepoSlug)

	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineVariableRepository(response)
}

func (r *Repository) AddPipelineKeyPair(rpkpo *RepositoryPipelineKeyPairOptions) (*PipelineKeyPair, error) {
	data := r.buildPipelineKeyPairBody(rpkpo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/ssh/key_pair", rpkpo.Owner, rpkpo.RepoSlug)

	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineKeyPairRepository(response)
}

func (r *Repository) UpdatePipelineBuildNumber(rpbno *RepositoryPipelineBuildNumberOptions) (*PipelineBuildNumber, error) {
	data := r.buildPipelineBuildNumberBody(rpbno)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/build_number", rpbno.Owner, rpbno.RepoSlug)

	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineBuildNumberRepository(response)
}

func (r *Repository) BranchingModel(rbmo *RepositoryBranchingModelOptions) (*BranchingModel, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/branching-model", rbmo.Owner, rbmo.RepoSlug)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeBranchingModel(response)
}

func (r *Repository) buildJsonBody(body map[string]interface{}) string {

	data, err := json.Marshal(body)
	if err != nil {
		pp.Println(err)
		os.Exit(9)
	}

	return string(data)
}

func (r *Repository) buildRepositoryBody(ro *RepositoryOptions) string {

	body := map[string]interface{}{}

	if ro.Scm != "" {
		body["scm"] = ro.Scm
	}
	//if ro.Scm != "" {
	//		body["name"] = ro.Name
	//}
	if ro.IsPrivate != "" {
		body["is_private"] = ro.IsPrivate
	}
	if ro.Description != "" {
		body["description"] = ro.Description
	}
	if ro.ForkPolicy != "" {
		body["fork_policy"] = ro.ForkPolicy
	}
	if ro.Language != "" {
		body["language"] = ro.Language
	}
	if ro.HasIssues != "" {
		body["has_issues"] = ro.HasIssues
	}
	if ro.HasWiki != "" {
		body["has_wiki"] = ro.HasWiki
	}
	if ro.Project != "" {
		body["project"] = map[string]string{
			"key": ro.Project,
		}
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineBody(rpo *RepositoryPipelineOptions) string {

	body := map[string]interface{}{}

	body["enabled"] = rpo.Enabled

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineVariableBody(rpvo *RepositoryPipelineVariableOptions) string {

	body := map[string]interface{}{}

	if rpvo.Uuid != "" {
		body["uuid"] = rpvo.Uuid
	}
	body["key"] = rpvo.Key
	body["value"] = rpvo.Value
	body["secured"] = rpvo.Secured

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineKeyPairBody(rpkpo *RepositoryPipelineKeyPairOptions) string {

	body := map[string]interface{}{}

	if rpkpo.PrivateKey != "" {
		body["private_key"] = rpkpo.PrivateKey
	}
	if rpkpo.PublicKey != "" {
		body["public_key"] = rpkpo.PublicKey
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineBuildNumberBody(rpbno *RepositoryPipelineBuildNumberOptions) string {

	body := map[string]interface{}{}

	body["next"] = rpbno.Next

	return r.buildJsonBody(body)
}

func decodeRepository(repoResponse interface{}) (*Repository, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var repository = new(Repository)
	err := mapstructure.Decode(repoMap, repository)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func decodeRepositoryFiles(repoResponse interface{}) ([]RepositoryFile, error) {
	repoFileMap := repoResponse.(map[string]interface{})

	if repoFileMap["type"] == "error" {
		return nil, DecodeError(repoFileMap)
	}

	var repositoryFiles = new([]RepositoryFile)
	err := mapstructure.Decode(repoFileMap["values"], repositoryFiles)
	if err != nil {
		return nil, err
	}

	return *repositoryFiles, nil
}

func decodeRepositoryBranches(branchResponse interface{}) (*RepositoryBranches, error) {

	var branchResponseMap map[string]interface{}
	err := json.Unmarshal(branchResponse.([]byte), &branchResponseMap)
	if err != nil {
		return nil, err
	}

	branchArray := branchResponseMap["values"].([]interface{})
	var branches []RepositoryBranch
	for _, branchEntry := range branchArray {
		var branch RepositoryBranch
		err = mapstructure.Decode(branchEntry, &branch)
		if err == nil {
			branches = append(branches, branch)
		}
	}

	page, ok := branchResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := branchResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := branchResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := branchResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := branchResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	repositoryBranches := RepositoryBranches{
		Page:     int(page),
		Pagelen:  int(pagelen),
		MaxDepth: int(max_depth),
		Size:     int(size),
		Next:     next,
		Branches: branches,
	}
	return &repositoryBranches, nil
}

func decodeRepositoryTags(tagResponse interface{}) (*RepositoryTags, error) {

	var tagResponseMap map[string]interface{}
	err := json.Unmarshal(tagResponse.([]byte), &tagResponseMap)
	if err != nil {
		return nil, err
	}

	tagArray := tagResponseMap["values"].([]interface{})
	var tags []RepositoryTag
	for _, tagEntry := range tagArray {
		var tag RepositoryTag
		err = mapstructure.Decode(tagEntry, &tag)
		if err == nil {
			tags = append(tags, tag)
		}
	}

	page, ok := tagResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := tagResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := tagResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := tagResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := tagResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	repositoryTags := RepositoryTags{
		Page:     int(page),
		Pagelen:  int(pagelen),
		MaxDepth: int(max_depth),
		Size:     int(size),
		Next:     next,
		Tags:     tags,
	}
	return &repositoryTags, nil
}

func decodePipelineRepository(repoResponse interface{}) (*Pipeline, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipeline = new(Pipeline)
	err := mapstructure.Decode(repoMap, pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

func decodePipelineVariableRepository(repoResponse interface{}) (*PipelineVariable, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineVariable = new(PipelineVariable)
	err := mapstructure.Decode(repoMap, pipelineVariable)
	if err != nil {
		return nil, err
	}

	return pipelineVariable, nil
}

func decodePipelineKeyPairRepository(repoResponse interface{}) (*PipelineKeyPair, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineKeyPair = new(PipelineKeyPair)
	err := mapstructure.Decode(repoMap, pipelineKeyPair)
	if err != nil {
		return nil, err
	}

	return pipelineKeyPair, nil
}

func decodePipelineBuildNumberRepository(repoResponse interface{}) (*PipelineBuildNumber, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineBuildNumber = new(PipelineBuildNumber)
	err := mapstructure.Decode(repoMap, pipelineBuildNumber)
	if err != nil {
		return nil, err
	}

	return pipelineBuildNumber, nil
}

func decodeBranchingModel(branchingModelResponse interface{}) (*BranchingModel, error) {
	branchingModelMap := branchingModelResponse.(map[string]interface{})

	if branchingModelMap["type"] == "error" {
		return nil, DecodeError(branchingModelMap)
	}

	var branchingModel = new(BranchingModel)
	err := mapstructure.Decode(branchingModelMap, branchingModel)
	if err != nil {
		return nil, err
	}

	return branchingModel, nil
}

func (rf RepositoryFile) String() string {
	return rf.Path
}

func (rb RepositoryBlob) String() string {
	return string(rb.Content)
}
