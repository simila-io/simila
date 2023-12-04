package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/acquirecloud/golibs/logging"
	"github.com/sashabaranov/go-openai"
	"io"
	"math"
	"os"
	"strings"
)

var log = logging.NewLogger("test")

func main() {
	client := openai.NewClient("la la la") // Simila

	//f, err := os.Open("/Users/dima/ai/xl-rg_1_en.txt")
	//if err != nil {
	//	log.Errorf("could not read file: %s", err)
	//	return
	//}
	//writeRecs("/Users/dima/ai/velocity.json", buildEmbedings(client, f))
	//f.Close()
	//return

	embs := readEmbeddings("/Users/dima/ai/velocity.json")
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter text: ")
		text, _ := reader.ReadString('\n')
		res := search(client, text, embs, 0.8)
		fmt.Println("I found this:")
		for i, r := range res {
			fmt.Printf("%d. %s\n", i+1, r)
		}
	}
}

type Rec struct {
	Text       string
	Embeddings []float32
}

func readEmbeddings(fn string) []Rec {
	f, err := os.Open(fn)
	if err != nil {
		log.Errorf("could not open data file %s: %s", fn, err)
		return nil
	}
	defer f.Close()

	d := json.NewDecoder(f)
	var res []Rec
	err = d.Decode(&res)
	if err != nil {
		log.Errorf("could not create the new decoder: %s", err)
		return nil
	}
	return res
}

func search(client *openai.Client, text string, data []Rec, th float32) []string {
	qr, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		log.Errorf("could not calculate embeddings: %s", err)
		return nil
	}
	e := qr.Data[0].Embedding
	m1, m2, m3 := float32(0), float32(0), float32(0)
	idx := []int{-1, -1, -1}
	for i, r := range data {
		s := cosineSim(e, r.Embeddings)
		if s < th {
			continue
		}
		if s > m1 {
			m2, m3 = m1, m2
			m1 = s
			idx[0], idx[1], idx[2] = i, idx[0], idx[1]
		} else if s > m2 {
			m3 = m2
			m2 = s
			idx[1], idx[2] = i, idx[1]
		} else if s > m3 {
			m3 = s
			idx[2] = i
		}
	}
	fmt.Println(m1)
	res := []string{}
	for _, v := range idx {
		if v > -1 {
			res = append(res, data[v].Text)
		}
	}
	return res
}

func calcSim(e1, e2 []float32) float32 {
	res := float32(0)
	for i, ev1 := range e1 {
		ev2 := e2[i]
		res += ev1 * ev2
	}
	return res
}

func cosineSim(a, b []float32) float32 {
	var sumA, s1, s2 float64
	for k := 0; k < len(a); k++ {
		sumA += float64(a[k] * b[k])
		s1 += float64(a[k] * a[k])
		s2 += float64(b[k] * b[k])
	}
	return float32(sumA / math.Sqrt(s1) * math.Sqrt(s2))
}

func buildEmbedings(client *openai.Client, f io.Reader) []Rec {
	log.Infof("start building embeddings")
	scanner := bufio.NewScanner(f)
	var recs []string
	res := []Rec{}
	recsLen := 0

	start := ""
	paragraph := 0
	for {
		var sb strings.Builder
		sb.WriteString(start)
		start = ""
		for scanner.Scan() {
			ln := scanner.Text()
			trimmed := strings.Trim(ln, " \t\n\v\f\r\x85\xA0")
			if len(trimmed) == 0 {
				// end of paragraph?
				if sb.Len() > 0 {
					break
				}
				continue
			}
			sb.WriteString(" ")
			sb.WriteString(trimmed)
			if sb.Len() > 8192 {
				s := sb.String()
				sb.Reset()
				idx := strings.LastIndex(s, ".")
				if idx > -1 {
					start = s[idx+1:]
					s = s[:idx+1]
				}
				sb.WriteString(s)
			}
		}
		if sb.Len() == 0 {
			break
		}
		paragraph++
		p := sb.String()
		recs = append(recs, p)
		recsLen += len(p)
		if recsLen >= 8192 {
			queryResponse, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
				Input: recs,
				Model: openai.AdaEmbeddingV2,
			})
			if err != nil {
				log.Errorf("could not create embeddings: %s", err)
				return nil
			}
			for i, e := range queryResponse.Data {
				res = append(res, Rec{Text: recs[i], Embeddings: e.Embedding})
			}
			recsLen = 0
			recs = recs[:0]
			log.Infof("just wrote new portion, the size so far is %d", len(res))
		}
	}

	if recsLen > 0 {
		queryResponse, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
			Input: recs,
			Model: openai.AdaEmbeddingV2,
		})
		if err != nil {
			log.Errorf("could not create embeddings: %s", err)
			return nil
		}
		for i, e := range queryResponse.Data {
			res = append(res, Rec{Text: recs[i], Embeddings: e.Embedding})
		}
		recsLen = 0
		recs = recs[:0]
	}

	log.Infof("completed %d records", len(res))

	return res

	//// Create an EmbeddingRequest for the user query
	//queryReq := openai.EmbeddingRequest{
	//	Input: []string{"How many chucks would a woodchuck chuck"},
	//	Model: openai.AdaEmbeddingV2,
	//}
	//
	//// Create an embedding for the user query
	//queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	//if err != nil {
	//	log.Errorf("Error creating query embedding: %s", err)
	//	return
	//}
	//
	//// Create an EmbeddingRequest for the target text
	//targetReq := openai.EmbeddingRequest{
	//	Input: []string{"How many chucks would a woodchuck chuck if the woodchuck could chuck wood"},
	//	Model: openai.AdaEmbeddingV2,
	//}
	//
	//// Create an embedding for the target text
	//targetResponse, err := client.CreateEmbeddings(context.Background(), targetReq)
	//if err != nil {
	//	log.Errorf("Error creating target embedding: %s", err)
	//	return
	//}
	//
	//// Now that we have the embeddings for the user query and the target text, we
	//// can calculate their similarity.
	//queryEmbedding := queryResponse.Data[0]
	//targetEmbedding := targetResponse.Data[0]
	//
	//similarity, err := queryEmbedding.DotProduct(&targetEmbedding)
	//if err != nil {
	//	log.Errorf("Error calculating dot product: %s", err)
	//	return
	//}
	//
	//log.Infof("The similarity score between the query and the target is %f", similarity)
}

func writeRecs(fn string, recs []Rec) {
	file, _ := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.Encode(recs)

	log.Infof("%d records written", len(recs))
}
