package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"net/http"
	"sync"
)

type data struct {
	Val mkbEntry `json:"value"`
}

type mkbEntry struct {
	Id          string
	CodeAndName string `json:"ninja_column_2"`
	LatinName   string `json:"ninja_column_3"`
}

func main() {
	allAvailableLinks := []string{
		"https://diagnosis.mk/wp-admin/admin-ajax.php?action=wp_ajax_ninja_tables_public_action&table_id=728&target_action=get-all-data&default_sorting=old_first&skip_rows=0&limit_rows=0&ninja_table_public_nonce=ffd33eb0fa&chunk_number=0",
		"https://diagnosis.mk/wp-admin/admin-ajax.php?action=wp_ajax_ninja_tables_public_action&table_id=728&target_action=get-all-data&default_sorting=old_first&skip_rows=0&limit_rows=0&chunk_number=1&ninja_table_public_nonce=ffd33eb0fa",
		"https://diagnosis.mk/wp-admin/admin-ajax.php?action=wp_ajax_ninja_tables_public_action&table_id=728&target_action=get-all-data&default_sorting=old_first&skip_rows=0&limit_rows=0&chunk_number=2&ninja_table_public_nonce=ffd33eb0fa",
		"https://diagnosis.mk/wp-admin/admin-ajax.php?action=wp_ajax_ninja_tables_public_action&table_id=728&target_action=get-all-data&default_sorting=old_first&skip_rows=0&limit_rows=0&chunk_number=3&ninja_table_public_nonce=ffd33eb0fa",
	}
	meilisearchClient := meilisearch.NewClient(meilisearch.ClientConfig{
		Host: "http://127.0.0.1:7700",
	})

	wg := sync.WaitGroup{}
	queue := make(chan struct{}, 20)

	for _, s := range allAvailableLinks {
		queue <- struct{}{}
		wg.Add(1)
		go scrape(s, meilisearchClient, &wg, queue)
	}

	wg.Wait()
	close(queue)
}

func scrape(url string, meilisearchClient *meilisearch.Client, wg *sync.WaitGroup, queue chan struct{}) {
	defer func() {
		<-queue
		wg.Done()
	}()

	//var mkbEntries []mkbEntry
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	decoded := json.NewDecoder(resp.Body)
	var res []data
	err = decoded.Decode(&res)
	if err != nil {
		panic(err)
	}

	saveToMeilisearch(res, meilisearchClient)
}

func saveToMeilisearch(values []data, meilisearchClient *meilisearch.Client) {
	tempMkbRegistry := meilisearchClient.Index("temp-mkb-registry")

	for _, value := range values {
		value.Val.Id = uuid.NewString()

		tempMkbRegistry.AddDocuments(value.Val)

	}
	fmt.Println("Done")
}
