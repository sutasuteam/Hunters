const express = require("express");
const fs = require("fs");
const path = require("path");
const multer = require("multer");

const router = express.Router();
const {
    createCanopyWallet,
    sendFaucetTx,
    sendTreasuryTx,
    sendLikePostTx,
    getAddressBalance
} = require("../utils/canopy");
// =========================
// PATH
// =========================
const USERS_FILE = path.join(__dirname, "../data/users.json");
const BALANCES_FILE = path.join(__dirname, "../data/balances.json");
const DATA_FILE = path.join(__dirname, "../data/tokens.json");
const TX_FILE = path.join(__dirname, "../data/transactions.json");
const POSTS_FILE = path.join(__dirname, "../data/posts.json");
const UPLOAD_DIR = path.join(__dirname, "../uploads/token");

// buat folder upload jika belum ada
if (!fs.existsSync(UPLOAD_DIR)) {
    fs.mkdirSync(UPLOAD_DIR, { recursive: true });
}

// =========================
// MULTER
// =========================

const storage = multer.diskStorage({

    destination(req, file, cb) {
        cb(null, UPLOAD_DIR);
    },

    filename(req, file, cb) {

        const ext = path.extname(file.originalname);

        const filename =
            Date.now() +
            "-" +
            Math.floor(Math.random() * 999999) +
            ext;

        cb(null, filename);

    }

});

const upload = multer({ storage });

// =========================
// JSON
// =========================

function loadTokens() {

    if (!fs.existsSync(DATA_FILE)) {

        fs.writeFileSync(DATA_FILE, "[]");

    }

    return JSON.parse(fs.readFileSync(DATA_FILE, "utf8"));

}


// get token

function getToken(tokenId) {
    return loadTokens().find(t => t.id === tokenId);
}


function saveTokens(tokens) {

    fs.writeFileSync(
        DATA_FILE,
        JSON.stringify(tokens, null, 2)
    );

}

function loadBalances() {

    if (!fs.existsSync(BALANCES_FILE)) {

        fs.writeFileSync(BALANCES_FILE, "[]");

    }

    return JSON.parse(
        fs.readFileSync(BALANCES_FILE, "utf8")
    );

}

function saveBalances(balances) {

    fs.writeFileSync(

        BALANCES_FILE,

        JSON.stringify(
            balances,
            null,
            2
        )

    );

}


function loadTransactions() {

    if (!fs.existsSync(TX_FILE)) {

        fs.writeFileSync(TX_FILE, "[]");

    }

    return JSON.parse(
        fs.readFileSync(TX_FILE, "utf8")
    );

}

function savePosts(posts) {

    fs.writeFileSync(

        POSTS_FILE,

        JSON.stringify(
            posts,
            null,
            2
        )

    );

}

function loadPosts() {

    if (!fs.existsSync(POSTS_FILE)) {
        fs.writeFileSync(POSTS_FILE, "[]");
    }

    return JSON.parse(
        fs.readFileSync(
            POSTS_FILE,
            "utf8"
        )
    );

}

function saveTransactions(tx) {

    fs.writeFileSync(

        TX_FILE,

        JSON.stringify(
            tx,
            null,
            2
        )

    );

}

router.get("/token-reserve/:tokenId", (req, res) => {

    try {

        const token =
            getToken(req.params.tokenId);

        if (!token) {

            return res.status(404).json({

                success: false,

                message: "Token not found"

            });

        }

        const reserve =
            getAddressBalance(
                token.treasuryAddress
            );

        res.json({

            success: true,

            tokenId: token.id,

            treasuryAddress:
                token.treasuryAddress,

            reserve:
                reserve.amount

        });

    } catch (err) {

        console.error(err);

        res.status(500).json({

            success: false,

            message: err.message

        });

    }

});

router.get("/my-tokens/:canopyAddress", (req, res) => {

    try {

        const canopyAddress =
            req.params.canopyAddress.toLowerCase();

        const users = loadUsers();

        const user = users.find(u =>

            (u.canopyAddress || "").toLowerCase() ===
            canopyAddress

        );

        if (!user) {

            return res.json({

                success: true,

                tokens: []

            });

        }

        const balances = loadBalances();

        const tokens = loadTokens();

        const myTokens = [];

        tokens.forEach(token => {
            console.log("TOKEN:");
console.log(token.name);

console.log("TREASURY:");
console.log(token.treasuryAddress);

            if (

                (token.creator || "").toLowerCase() !==
                user.canopyAddress.toLowerCase()

            ) {
                return;
            }

            const balance = balances.find(b =>

                b.tokenId === token.id &&
                b.wallet.toLowerCase() ===
                user.canopyAddress.toLowerCase()

            );

            let reserve = 0;

            console.log("CALLING RPC...");
            
            try {
            
                const chainReserve =
                    getAddressBalance(
                        token.treasuryAddress
                    );
            
                console.log("RPC RESULT:");
                console.log(chainReserve);
            
                reserve =
                    Number(chainReserve.amount || 0);
            
            } catch (e) {
            
                console.log(
                    "Reserve Error:",
                    e.message
                );
            
            }

            myTokens.push({

                id: token.id,

                creator: token.creator,

                name: token.name,

                symbol: token.symbol,

                logo: token.logo,

                description: token.description,

                treasuryAddress:
                    token.treasuryAddress,

                treasuryNickname:
                    token.treasuryNickname,

                faucetTx:
                    token.faucetTx,

                createdAt:
                    token.createdAt,

                supply:
                    token.supply,

                balance:
                    balance
                        ? balance.balance
                        : 0,

                reserve

            });

        });

        console.log("=== MY TOKENS ===");
console.dir(myTokens, { depth: null });

        res.json({

            success: true,

            tokens: myTokens

        });

    } catch (err) {

        console.error(err);

        res.status(500).json({

            success: false,

            message: err.message

        });

    }

});

function loadUsers() {

    if (!fs.existsSync(USERS_FILE)) {
        fs.writeFileSync(USERS_FILE, "[]");
    }

    return JSON.parse(
        fs.readFileSync(
            USERS_FILE,
            "utf8"
        )
    );

}

// =========================
// CREATE TOKEN
// =========================
router.post(
    "/create-token",
    upload.single("logo"),
    (req, res) => {

        console.log("========== CREATE TOKEN ==========");
        console.log("BODY:", req.body);
        console.log("FILE:", req.file);

        try {

            const {
                creator,
                name,
                symbol,
                supply,
                description
            } = req.body;

            console.log("creator =", creator);
            console.log("name =", name);
            console.log("symbol =", symbol);
            console.log("supply =", supply);

            if (!creator || !name || !symbol || !supply) {

                return res.status(400).json({
                    success: false,
                    message: "Missing required fields"
                });

            }

            const tokens = loadTokens();

            console.log("TOKENS =", tokens);

            const exist = tokens.find(
                t => t.symbol.toUpperCase() === symbol.toUpperCase()
            );

            if (exist) {

                return res.status(400).json({
                    success: false,
                    message: "Token symbol already exists"
                });

            }

            let logo = "";

            if (req.file) {
                logo = "/uploads/token/" + req.file.filename;
            }


            const treasury = createCanopyWallet();

const faucet = sendFaucetTx(treasury.address);

if (!faucet.success) {

    return res.status(500).json({

        success:false,

        message:faucet.error

    });

}
const token = {

    id: Date.now().toString(),

    creator,

    name,

    symbol: symbol.toUpperCase(),

    supply: Number(supply),

    treasuryNickname: treasury.nickname,

    treasuryAddress: treasury.address,

    faucetTx: faucet.txHash,

    logo,

    description: description || "",

    createdAt: new Date().toISOString()

};
            tokens.push(token);

            saveTokens(tokens);


// ======================================
// Buat saldo awal creator
// ======================================

const balances = loadBalances();

// kalau token ini belum pernah punya balance
const exists = balances.find(b =>

    b.tokenId === token.id &&
    b.wallet.toLowerCase() === creator.toLowerCase()

);

if (!exists) {

    balances.push({

        tokenId: token.id,

        wallet: creator,

        balance: Number(supply),

        updatedAt: new Date().toISOString()

    });

    saveBalances(balances);

}

            console.log("TOKEN SAVED:", token);

            console.log("BALANCE CREATED:", {

                tokenId: token.id,
            
                wallet: creator,
            
                balance: Number(supply)
            
            });

            return res.json({
                success: true,
                token
            });

        } catch (err) {

            console.error("CREATE TOKEN ERROR");
            console.error(err);

            return res.status(500).json({
                success: false,
                message: err.message
            });

        }

    }
);


router.post("/like-post", (req, res) => {

    try {

        const {
            postId,
            liker
        } = req.body;

        if (!postId || !liker) {

            return res.status(400).json({
                success: false,
                message: "Missing required fields"
            });

        }

        const posts = loadPosts();

        const post = posts.find(
            p => String(p.id) === String(postId)
        );

        if (!post) {

            return res.status(404).json({
                success: false,
                message: "Post not found"
            });

        }

        const users = loadUsers();

        const ownerUser = users.find(u =>

            (u.ethAddress || "").toLowerCase() ===
            (post.owner || "").toLowerCase()
        
        );
        if (!ownerUser) {

            return res.status(404).json({
                success: false,
                message: "Post owner not found"
            });

        }

        const liked = !post.liked;

        const tx = sendLikePostTx(

            liker,

            ownerUser.canopyAddress,

            Number(post.id),

            liked

        );

        if (!tx.success) {

            return res.status(400).json({

                success: false,

                message: tx.error

            });

        }

        post.liked = liked;

        if (liked) {

            post.likes = (post.likes || 0) + 1;

        } else {

            post.likes = Math.max(
                (post.likes || 0) - 1,
                0
            );

        }

        savePosts(posts);

        res.json({

            success: true,

            liked,

            likes: post.likes,

            txHash: tx.txHash

        });

    }

    catch (err) {

        console.error(err);

        res.status(500).json({

            success: false,

            message: err.message

        });

    }

});


router.post("/send-token", (req, res) => {

    console.log("========== SEND TOKEN ==========");
    console.log(req.body);


    try {

        const {

            tokenId,
            from,
            to,
            amount

        } = req.body;

        if (

            !tokenId ||
            !from ||
            !to ||
            !amount

        ) {

            return res.status(400).json({

                success: false,

                message: "Missing required fields"

            });

        }

        const sendAmount = Number(amount);

        if (sendAmount <= 0) {

            return res.status(400).json({

                success: false,

                message: "Invalid amount"

            });

        }

        // ===========================
        // Load Token
        // ===========================

        const tokens = loadTokens();

        const token = tokens.find(

            t => t.id === tokenId

        );

        if (!token) {

            return res.status(404).json({

                success: false,

                message: "Token not found"

            });

        }

        // ===========================
        // Cek Reserve Treasury
        // ===========================

        const chainReserve = getAddressBalance(
            token.treasuryAddress
        );

        const reserve = Number(chainReserve.amount);

        if (reserve < sendAmount) {

            return res.status(400).json({

                success: false,

                message: "Reserve is not enough"

            });

        }

        // ===========================
        // Kirim dari Treasury Wallet
        // ===========================

        const treasuryTx = sendTreasuryTx(

            token.treasuryNickname,
        
            to,
        
            sendAmount
        
        );

        if (!treasuryTx.success) {

            return res.status(500).json({

                success: false,

                message: "Failed sending treasury token"

            });

        }

        // ===========================
        // Update Receiver Balance
        // ===========================

        const balances = loadBalances();

        let receiver = balances.find(b =>

            b.tokenId === tokenId &&
            b.wallet.toLowerCase() === to.toLowerCase()

        );

        if (receiver) {

            receiver.balance += sendAmount;

            receiver.updatedAt =
                new Date().toISOString();

        }

        else {

            receiver = {

                tokenId,

                wallet: to,

                balance: sendAmount,

                updatedAt:
                    new Date().toISOString()

            };

            balances.push(receiver);

        }

        saveBalances(balances);

        // ===========================
        // Save Transaction History
        // ===========================

        const transactions = loadTransactions();

        transactions.unshift({

            id: Date.now().toString(),

            tokenId,

            from,

            to,

            amount: sendAmount,

            txHash: treasuryTx.txHash,

            timestamp:
                new Date().toISOString()

        });

        saveTransactions(transactions);

        // ===========================
        // Ambil Reserve Terbaru
        // ===========================

        const latestReserve = getAddressBalance(
            token.treasuryAddress
        );

        res.json({

            success: true,

            message: "Transfer success",

            reserve: Number(latestReserve.amount),

            receiverBalance: receiver.balance,

            txHash: treasuryTx.txHash

        });

    }

    catch (err) {

        console.error(err);

        res.status(500).json({

            success: false,

            message: err.message

        });

    }

});

module.exports = router;


