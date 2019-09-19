package cosmosdb

import (
	"net/http"
)

// Collection represents a collection
type Collection struct {
	ID                       string                    `json:"id,omitempty"`
	ResourceID               string                    `json:"_rid,omitempty"`
	Timestamp                int                       `json:"_ts,omitempty"`
	Self                     string                    `json:"_self,omitempty"`
	ETag                     string                    `json:"_etag,omitempty"`
	Documents                string                    `json:"_docs,omitempty"`
	StoredProcedures         string                    `json:"_sprocs,omitempty"`
	Triggers                 string                    `json:"_triggers,omitempty"`
	UserDefinedFunctions     string                    `json:"_udfs,omitempty"`
	Conflicts                string                    `json:"_conflicts,omitempty"`
	IndexingPolicy           *IndexingPolicy           `json:"indexingPolicy,omitempty"`
	PartitionKey             *PartitionKey             `json:"partitionKey,omitempty"`
	ConflictResolutionPolicy *ConflictResolutionPolicy `json:"conflictResolutionPolicy,omitempty"`
	GeospatialConfig         *GeospatialConfig         `json:"geospatialConfig,omitempty"`
}

// IndexingPolicy represents an indexing policy
type IndexingPolicy struct {
	Automatic     bool           `json:"automatic,omitempty"`
	IndexingMode  string         `json:"indexingMode,omitempty"`
	IncludedPaths []IncludedPath `json:"includedPaths,omitempty"`
	ExcludedPaths []IncludedPath `json:"excludedPaths,omitempty"`
}

// IncludedPath represents an included path
type IncludedPath struct {
	Path    string  `json:"path,omitempty"`
	Indexes []Index `json:"indexes,omitempty"`
}

// Index represents an index
type Index struct {
	DataType  string `json:"dataType,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Precision int    `json:"precision,omitempty"`
}

// ExcludedPath represents an excluded path
type ExcludedPath struct {
	Path string `json:"path,omitempty"`
}

// PartitionKey represents a partition key
type PartitionKey struct {
	Paths   []string `json:"paths,omitempty"`
	Kind    string   `json:"kind,omitempty"`
	Version int      `json:"version,omitempty"`
}

// ConflictResolutionPolicy represents a conflict resolution policy
type ConflictResolutionPolicy struct {
	Mode                        string `json:"mode,omitempty"`
	ConflictResolutionPath      string `json:"conflictResolutionPath,omitempty"`
	ConflictResolutionProcedure string `json:"conflictResolutionProcedure,omitempty"`
}

// GeospatialConfig represents a geospatial config
type GeospatialConfig struct {
	Type string `json:"type,omitempty"`
}

// Collections represents collections
type Collections struct {
	Count       int          `json:"_count,omitempty"`
	ResourceID  string       `json:"_rid,omitempty"`
	Collections []Collection `json:"DocumentCollections,omitempty"`
}

// PartitionKeyRanges represents partition key ranges
type PartitionKeyRanges struct {
	Count              int                 `json:"_count,omitempty"`
	ResourceID         string              `json:"_rid,omitempty"`
	PartitionKeyRanges []PartitionKeyRange `json:"PartitionKeyRanges,omitempty"`
}

// PartitionKeyRange represents a partition key range
type PartitionKeyRange struct {
	ID                 string   `json:"id,omitempty"`
	ResourceID         string   `json:"_rid,omitempty"`
	Timestamp          int      `json:"_ts,omitempty"`
	Self               string   `json:"_self,omitempty"`
	ETag               string   `json:"_etag,omitempty"`
	MaxExclusive       string   `json:"maxExclusive,omitempty"`
	MinInclusive       string   `json:"minInclusive,omitempty"`
	ResourceIDPrefix   int      `json:"ridPrefix,omitempty"`
	ThroughputFraction int      `json:"throughputFraction,omitempty"`
	Status             string   `json:"status,omitempty"`
	Parents            []string `json:"parents,omitempty"`
}

type collectionClient struct {
	*databaseClient
	path string
}

// CollectionClient is a collection client
type CollectionClient interface {
	Create(*Collection) (*Collection, error)
	List() CollectionIterator
	Get(string) (*Collection, error)
	Delete(*Collection) error
	Replace(*Collection) (*Collection, error)
	PartitionKeyRanges(string) (*PartitionKeyRanges, error)
}

type collectionListIterator struct {
	*collectionClient
	continuation string
	done         bool
}

// CollectionIterator is a collection iterator
type CollectionIterator interface {
	Next() (*Collections, error)
}

// NewCollectionClient returns a new collection client
func NewCollectionClient(c DatabaseClient, dbid string) CollectionClient {
	return &collectionClient{
		databaseClient: c.(*databaseClient),
		path:           "dbs/" + dbid,
	}
}

func (c *collectionClient) Create(newcoll *Collection) (coll *Collection, err error) {
	err = c.do(http.MethodPost, c.path+"/colls", "colls", c.path, http.StatusCreated, &newcoll, &coll, nil)
	return
}

func (c *collectionClient) List() CollectionIterator {
	return &collectionListIterator{collectionClient: c}
}

func (c *collectionClient) Get(collid string) (coll *Collection, err error) {
	err = c.do(http.MethodGet, c.path+"/colls/"+collid, "colls", c.path+"/colls/"+collid, http.StatusOK, nil, &coll, nil)
	return
}

func (c *collectionClient) Delete(coll *Collection) error {
	if coll.ETag == "" {
		return ErrETagRequired
	}
	headers := http.Header{}
	headers.Set("If-Match", coll.ETag)
	return c.do(http.MethodDelete, c.path+"/colls/"+coll.ID, "colls", c.path+"/colls/"+coll.ID, http.StatusNoContent, nil, nil, headers)
}

func (c *collectionClient) Replace(newcoll *Collection) (coll *Collection, err error) {
	err = c.do(http.MethodPost, c.path+"/colls/"+newcoll.ID, "colls", c.path+"/colls/"+newcoll.ID, http.StatusCreated, &newcoll, &coll, nil)
	return
}

func (c *collectionClient) PartitionKeyRanges(collid string) (pkrs *PartitionKeyRanges, err error) {
	err = c.do(http.MethodGet, c.path+"/colls/"+collid+"/pkranges", "pkranges", c.path+"/colls/"+collid, http.StatusOK, nil, &pkrs, nil)
	return
}

func (i *collectionListIterator) Next() (colls *Collections, err error) {
	if i.done {
		return
	}

	headers := http.Header{}
	if i.continuation != "" {
		headers.Set("X-Ms-Continuation", i.continuation)
	}

	err = i.do(http.MethodGet, i.path+"/colls", "colls", i.path, http.StatusOK, nil, &colls, headers)
	if err != nil {
		return
	}

	i.continuation = headers.Get("X-Ms-Continuation")
	i.done = i.continuation == ""

	return
}