package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
"github.com/joho/godotenv"
"github.com/canopy-network/go-plugin/contract"
"github.com/canopy-network/go-plugin/crypto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	defaultPublicRPC = "http://127.0.0.1:50002"
	defaultAdminRPC  = "http://127.0.0.1:50003"
	defaultSigner    = "validator"
	defaultPassword  = ""
	defaultFee       = uint64(10000)
	defaultNetworkID = uint64(1)
	defaultChainID   = uint64(1)
)

const (
    balancesFile  = "../../../data/balances.json"
    transfersFile = "../../../data/transfers.json"
)

type MintRequest struct {
	Creator  string `json:"creator"`
	TokenID  string `json:"tokenId"`
	TokenID2 string `json:"token_id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	Metadata string `json:"metadata"`
	URI      string `json:"uri"`
}

type Balance struct {
    TokenID   string `json:"tokenId"`
    Wallet    string `json:"wallet"`
    Balance   uint64 `json:"balance"`
    UpdatedAt string `json:"updatedAt"`
}

type Transfer struct {
    TxID      string `json:"txId"`
    TokenID   string `json:"tokenId"`
    From      string `json:"from"`
    To        string `json:"to"`
    Amount    uint64 `json:"amount"`
    Timestamp uint64 `json:"timestamp"`
}


type StakeRequest struct {
    TokenID string `json:"tokenId"`
    Stake   bool   `json:"stake"`
}


type LikePostRequest struct {
    Address  string `json:"address"`
    Nickname string `json:"nickname"`
    Password string `json:"password"`

    Owner  string `json:"owner"`
    PostID uint64 `json:"postId"`
    Liked  bool   `json:"liked"`
}

type keyGroup struct {
	Address    string `json:"address"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}


func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}


// ===========================================
// STORAGE
// ===========================================

func loadBalances() ([]Balance, error) {

    var balances []Balance

    data, err := os.ReadFile(balancesFile)
    if err != nil {
        return balances, err
    }

    err = json.Unmarshal(data, &balances)

    return balances, err
}

func saveBalances(balances []Balance) error {

    data, err := json.MarshalIndent(
        balances,
        "",
        "  ",
    )

    if err != nil {
        return err
    }

    return os.WriteFile(
        balancesFile,
        data,
        0644,
    )
}

func loadTransfers() ([]Transfer, error) {

    var transfers []Transfer

    data, err := os.ReadFile(transfersFile)
    if err != nil {
        return transfers, err
    }

    err = json.Unmarshal(data, &transfers)

    return transfers, err
}

func saveTransfers(transfers []Transfer) error {

    data, err := json.MarshalIndent(
        transfers,
        "",
        "  ",
    )

    if err != nil {
        return err
    }

    return os.WriteFile(
        transfersFile,
        data,
        0644,
    )
}

func findBalance(
    balances []Balance,
    wallet string,
    tokenID string,
) *Balance {

    for i := range balances {

        if balances[i].Wallet == wallet &&
            balances[i].TokenID == tokenID {

            return &balances[i]

        }

    }

    return nil
}


func main() {
	godotenv.Load()
	fmt.Println("CANOPY_SIGNER_ADDRESS =", os.Getenv("CANOPY_SIGNER_ADDRESS"))
    fmt.Println("CANOPY_PASSWORD =", os.Getenv("CANOPY_PASSWORD"))
	http.Handle("/", http.FileServer(http.Dir("./public")))

	fmt.Println("========== MINT REQUEST ==========")


	http.HandleFunc("/mint", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST /mint", http.StatusMethodNotAllowed)
			return
		}

		var req MintRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"status": "error",
				"error":  "invalid JSON: " + err.Error(),
			})
			return
		}
		fmt.Printf("%+v\n", req)

		publicRPC := envString("CANOPY_RPC", defaultPublicRPC)
		adminRPC := envString("CANOPY_ADMIN_RPC", defaultAdminRPC)
		signerAddress := envString("CANOPY_SIGNER_ADDRESS", defaultSigner)
		password := envString("CANOPY_PASSWORD", defaultPassword)
		fee := envUint64("CANOPY_FEE", defaultFee)
		networkID := envUint64("CANOPY_NETWORK_ID", defaultNetworkID)
		chainID := envUint64("CANOPY_CHAIN_ID", defaultChainID)

		tokenID := req.TokenID
		if tokenID == "" {
			tokenID = req.TokenID2
		}
		if tokenID == "" {
			tokenID = "inkarnasi-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
		}

		name := req.Name
		if name == "" {
			name = "Inkarnasi"
		}

		image := req.Image
		if image == "" {
			image = req.URI
		}
		if image == "" {
			image = "http://localhost:3001/images/inkarnasi.png"
		}

		metadata := req.Metadata

		if metadata == "" {
		
			metadata = "{}"
		
		}

		result, err := mintNFT(publicRPC, adminRPC, signerAddress, password, tokenID, name, image, metadata, fee, networkID, chainID)
		if err != nil {
			log.Println("MINT ERROR:", err)
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "success",
			"tokenId": tokenID,
			"rpc":     result,
		})
	})

	http.HandleFunc("/stake", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Use POST /stake", http.StatusMethodNotAllowed)
			return
		}
	
		var req StakeRequest
	
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	
			writeJSON(w, 400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
	
			return
		}
	
		result, err := setStakeNFT(
	
			envString("CANOPY_RPC", defaultPublicRPC),
	
			envString("CANOPY_ADMIN_RPC", defaultAdminRPC),
	
			envString("CANOPY_SIGNER_ADDRESS", defaultSigner),
	
			envString("CANOPY_PASSWORD", defaultPassword),
	
			req.TokenID,
	
			true,
	
			defaultFee,
	
			defaultNetworkID,
	
			defaultChainID,
	
		)
	
		if err != nil {
	
			writeJSON(w, 500, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
	
			return
		}
	
		writeJSON(w, 200, map[string]interface{}{
			"success": true,
			"result":  result,
		})
	
	})

	http.HandleFunc("/unstake", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Use POST /unstake", http.StatusMethodNotAllowed)
			return
		}
	
		var req StakeRequest
	
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	
			writeJSON(w, 400, map[string]interface{}{
				"success": false,
				"error": err.Error(),
			})
	
			return
		}
	
		result, err := setStakeNFT(
	
			envString("CANOPY_RPC", defaultPublicRPC),
	
			envString("CANOPY_ADMIN_RPC", defaultAdminRPC),
	
			envString("CANOPY_SIGNER_ADDRESS", defaultSigner),
	
			envString("CANOPY_PASSWORD", defaultPassword),
	
			req.TokenID,
	
			false,
	
			defaultFee,
	
			defaultNetworkID,
	
			defaultChainID,
	
		)
	
		if err != nil {
	
			writeJSON(w,500,map[string]interface{}{
				"success":false,
				"error":err.Error(),
			})
	
			return
		}
	
		writeJSON(w,200,map[string]interface{}{
			"success":true,
			"result":result,
		})
	
	})

	http.HandleFunc("/like-post", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Use POST /like-post", http.StatusMethodNotAllowed)
			return
		}
	
		var req LikePostRequest
	
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	
			writeJSON(w, 400, map[string]interface{}{
				"success": false,
				"error": err.Error(),
			})
	
			return
		}
	
		result, err := likePost(
	
			envString("CANOPY_RPC", defaultPublicRPC),
	
			envString("CANOPY_ADMIN_RPC", defaultAdminRPC),
	
			envString("CANOPY_SIGNER_ADDRESS", defaultSigner),
	
			envString("CANOPY_PASSWORD", defaultPassword),
	
			req.Owner,
	
			req.PostID,
	
			req.Liked,
	
			defaultFee,
	
			defaultNetworkID,
	
			defaultChainID,
		)
	
		if err != nil {
	
			writeJSON(w, 500, map[string]interface{}{
				"success": false,
				"error": err.Error(),
			})
	
			return
		}
	
		writeJSON(w, 200, map[string]interface{}{
			"success": true,
			"result": result,
		})
	
	})

	fmt.Println("🚀 Mint signer jalan di http://localhost:8081")
	fmt.Println("RPC:", envString("CANOPY_RPC", defaultPublicRPC))
	fmt.Println("Admin RPC:", envString("CANOPY_ADMIN_RPC", defaultAdminRPC))
	fmt.Println("Signer:", envString("CANOPY_SIGNER_ADDRESS", defaultSigner))
	log.Fatal(
		http.ListenAndServe(
			":8081",
			enableCORS(http.DefaultServeMux),
		),
	)
}

func mintNFT(
    publicRPC string,
    adminRPC string,
    signerAddress string,
    password string,
	tokenID string,
	name string,
	image string,
	metadata string,
	fee uint64,
	networkID uint64,
	chainID uint64,
) (map[string]interface{}, error) {
	height, err := getHeight(publicRPC)
	fmt.Println("HEIGHT =", height)
	if err != nil {
		return nil, fmt.Errorf("get height failed: %w", err)
	}

	signerKey, err := keystoreGetKey(
		adminRPC,
		signerAddress,
		password,
	)

	fmt.Printf("SIGNER = %+v\n", signerKey)
	if err != nil {
		return nil, fmt.Errorf("keystore get key failed: %w", err)
	}

	creatorBytes, err := hex.DecodeString(signerKey.Address)
	if err != nil {
		return nil, fmt.Errorf("decode creator address failed: %w", err)
	}
	if len(creatorBytes) != 20 {
		return nil, fmt.Errorf("creator address must be 20 bytes, got %d", len(creatorBytes))
	}

	msg := &contract.MessageMintNFT{
		Creator:  creatorBytes,
		TokenId:  tokenID,
		Name:     name,
		Image:    image,
		Metadata: metadata,
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal MessageMintNFT failed: %w", err)
	}

	typeURL := "type.googleapis.com/types.MessageMintNFT"
	msgType := "mint_nft"
	msgAny := &anypb.Any{
		TypeUrl: typeURL,
		Value:   msgBytes,
	}

	msgJSONBytes, err := protojson.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("protojson marshal msg failed: %w", err)
	}

	var msgJSON map[string]interface{}
	if err := json.Unmarshal(msgJSONBytes, &msgJSON); err != nil {
		return nil, fmt.Errorf("json unmarshal msg failed: %w", err)
	}

	txTime := uint64(time.Now().UnixNano())

	signBytes, err := crypto.GetSignBytes(
		msgType,
		msgAny,
		txTime,
		height,
		fee,
		"",
		networkID,
		chainID,
	)
	if err != nil {
		return nil, fmt.Errorf("get sign bytes failed: %w", err)
	}

	privKey, err := crypto.StringToBLS12381PrivateKey(signerKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key failed: %w", err)
	}

	signature := privKey.Sign(signBytes)

	pubKeyBytes, err := hex.DecodeString(signerKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("decode public key failed: %w", err)
	}

	tx := map[string]interface{}{
		"type":       msgType,
		"msgTypeUrl": typeURL,
		"msg":        msgJSON,
		"signature": map[string]string{
			"publicKey": hex.EncodeToString(pubKeyBytes),
			"signature": hex.EncodeToString(signature),
		},
		"time":          txTime,
		"createdHeight": height,
		"fee":           fee,
		"memo":          "",
		"networkID":     networkID,
		"chainID":       chainID,
	}

	txBody, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("marshal tx JSON failed: %w", err)
	}

	respBody, err := postRawJSON(publicRPC+"/v1/tx", string(txBody))

	fmt.Println(string(respBody))

	if err != nil {
		return nil, fmt.Errorf("submit tx failed: %w", err)
	}




	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		result = map[string]interface{}{
			"raw": string(respBody),
		}
	}

	return result, nil
}

func setStakeNFT(

	publicRPC string,

	adminRPC string,

	signerAddress string,

	password string,

	tokenID string,

	stake bool,

	fee uint64,

	networkID uint64,

	chainID uint64,

) (map[string]interface{}, error) {

	height, err := getHeight(publicRPC)
	if err != nil {
		return nil, fmt.Errorf("get height failed: %w", err)
	}

	signerKey, err := keystoreGetKey(
		adminRPC,
		signerAddress,
		password,
	)
	if err != nil {
		return nil, fmt.Errorf("keystore get key failed: %w", err)
	}

	ownerBytes, err := hex.DecodeString(signerKey.Address)
	if err != nil {
		return nil, fmt.Errorf("decode owner address failed: %w", err)
	}

	if len(ownerBytes) != 20 {
		return nil, fmt.Errorf(
			"owner address must be 20 bytes, got %d",
			len(ownerBytes),
		)
	}

	msg := &contract.MessageSetStake{
		Owner:   ownerBytes,
		TokenId: tokenID,
		Stake:   stake,
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal MessageSetStake failed: %w",
			err,
		)
	}

	typeURL := "type.googleapis.com/types.MessageSetStake"
	msgType := "set_stake"

	msgAny := &anypb.Any{
		TypeUrl: typeURL,
		Value:   msgBytes,
	}

	msgJSONBytes, err := protojson.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf(
			"protojson marshal failed: %w",
			err,
		)
	}

	var msgJSON map[string]interface{}

	if err := json.Unmarshal(msgJSONBytes, &msgJSON); err != nil {
		return nil, fmt.Errorf(
			"json unmarshal failed: %w",
			err,
		)
	}

	txTime := uint64(time.Now().UnixNano())

	signBytes, err := crypto.GetSignBytes(
		msgType,
		msgAny,
		txTime,
		height,
		fee,
		"",
		networkID,
		chainID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"get sign bytes failed: %w",
			err,
		)
	}

	privKey, err := crypto.StringToBLS12381PrivateKey(
		signerKey.PrivateKey,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"parse private key failed: %w",
			err,
		)
	}

	signature := privKey.Sign(signBytes)

	pubKeyBytes, err := hex.DecodeString(
		signerKey.PublicKey,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"decode public key failed: %w",
			err,
		)
	}

	tx := map[string]interface{}{
		"type":       msgType,
		"msgTypeUrl": typeURL,
		"msg":        msgJSON,
		"signature": map[string]string{
			"publicKey": hex.EncodeToString(pubKeyBytes),
			"signature": hex.EncodeToString(signature),
		},
		"time":          txTime,
		"createdHeight": height,
		"fee":           fee,
		"memo":          "",
		"networkID":     networkID,
		"chainID":       chainID,
	}

	txBody, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal tx JSON failed: %w",
			err,
		)
	}

	respBody, err := postRawJSON(
		publicRPC+"/v1/tx",
		string(txBody),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"submit tx failed: %w",
			err,
		)
	}

	var result map[string]interface{}

	if err := json.Unmarshal(respBody, &result); err != nil {

		result = map[string]interface{}{
			"raw": string(respBody),
		}

	}

	return result, nil
}


func likePost(

	publicRPC string,

	adminRPC string,

	signerAddress string,

	password string,

	owner string,

	postID uint64,

	liked bool,

	fee uint64,

	networkID uint64,

	chainID uint64,

) (map[string]interface{}, error) {

	height, err := getHeight(publicRPC)
	if err != nil {
		return nil, fmt.Errorf("get height failed: %w", err)
	}

	signerKey, err := keystoreGetKey(
		adminRPC,
		signerAddress,
		password,
	)

	if err != nil {
		return nil, fmt.Errorf("keystore get key failed: %w", err)
	}

	likerBytes, err := hex.DecodeString(signerKey.Address)
	if err != nil {
		return nil, fmt.Errorf("decode liker address failed: %w", err)
	}

	if len(likerBytes) != 20 {
		return nil, fmt.Errorf(
			"liker address must be 20 bytes, got %d",
			len(likerBytes),
		)
	}

	ownerBytes, err := hex.DecodeString(owner)
	if err != nil {
		return nil, fmt.Errorf("decode owner address failed: %w", err)
	}

	if len(ownerBytes) != 20 {
		return nil, fmt.Errorf(
			"owner address must be 20 bytes, got %d",
			len(ownerBytes),
		)
	}

	msg := &contract.MessageLikePost{

		Liker: likerBytes,

		Owner: ownerBytes,

		PostId: postID,

		Liked: liked,
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal MessageLikePost failed: %w",
			err,
		)
	}

	typeURL := "type.googleapis.com/contract.MessageLikePost"

	msgType := "like_post"

	msgAny := &anypb.Any{
		TypeUrl: typeURL,
		Value:   msgBytes,
	}

	msgJSONBytes, err := protojson.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf(
			"protojson marshal failed: %w",
			err,
		)
	}

	var msgJSON map[string]interface{}

	if err := json.Unmarshal(msgJSONBytes, &msgJSON); err != nil {
		return nil, fmt.Errorf(
			"json unmarshal failed: %w",
			err,
		)
	}

	txTime := uint64(time.Now().UnixNano())

	signBytes, err := crypto.GetSignBytes(

		msgType,

		msgAny,

		txTime,

		height,

		fee,

		"",

		networkID,

		chainID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"get sign bytes failed: %w",
			err,
		)
	}

	privKey, err := crypto.StringToBLS12381PrivateKey(
		signerKey.PrivateKey,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"parse private key failed: %w",
			err,
		)
	}

	signature := privKey.Sign(signBytes)

	pubKeyBytes, err := hex.DecodeString(
		signerKey.PublicKey,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"decode public key failed: %w",
			err,
		)
	}

	tx := map[string]interface{}{

		"type": msgType,

		"msgTypeUrl": typeURL,

		"msg": msgJSON,

		"signature": map[string]string{

			"publicKey": hex.EncodeToString(pubKeyBytes),

			"signature": hex.EncodeToString(signature),

		},

		"time": txTime,

		"createdHeight": height,

		"fee": fee,

		"memo": "",

		"networkID": networkID,

		"chainID": chainID,
	}

	txBody, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal tx JSON failed: %w",
			err,
		)
	}

	respBody, err := postRawJSON(

		publicRPC+"/v1/tx",

		string(txBody),

	)

	if err != nil {
		return nil, fmt.Errorf(
			"submit tx failed: %w",
			err,
		)
	}

	var result map[string]interface{}

	if err := json.Unmarshal(respBody, &result); err != nil {

		result = map[string]interface{}{
			"raw": string(respBody),
		}

	}

	return result, nil

}

func keystoreGetKey(rpcURL, addressOrNickname, password string) (*keyGroup, error) {
	reqJSON := fmt.Sprintf(
		`{"address":"%s","password":"%s"}`,
		addressOrNickname,
		password,
	)

	respBody, err := postRawJSON(rpcURL+"/v1/admin/keystore-get", reqJSON)
	if err != nil {
		return nil, err
	}

	var kg keyGroup
	if err := json.Unmarshal(respBody, &kg); err != nil {
		return nil, fmt.Errorf("failed to parse keystore response: %v, body: %s", err, string(respBody))
	}

	if kg.Address == "" || kg.PrivateKey == "" || kg.PublicKey == "" {
		return nil, fmt.Errorf("keystore-get returned incomplete key: %s", string(respBody))
	}

	return &kg, nil
}

func getHeight(rpcURL string) (uint64, error) {
	respBody, err := postRawJSON(rpcURL+"/v1/query/height", `{}`)
	if err != nil {
		return 0, err
	}

	var result struct {
		Height uint64 `json:"height"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("failed to parse height response: %v, body: %s", err, string(respBody))
	}

	return result.Height, nil
}

func postRawJSON(url string, reqJSON string) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(reqJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, url, string(body))
	}

	return body, nil
}

func writeJSON(w http.ResponseWriter, status int, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func envString(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func envUint64(key string, fallback uint64) uint64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}