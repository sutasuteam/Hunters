require("dotenv").config();

const express = require("express");
const cors = require("cors");
const fs = require("fs");
const path = require("path");
const multer = require("multer");
const { execSync } = require("child_process");
const { execFileSync } = require("child_process");




console.log("POST JSON SAVED");

console.log("BROWSER RELOAD");

const PORT = process.env.PORT || 3001;

const POSTS_FILE =
path.join(__dirname,"data","posts.json");

const DB_FILE =
path.join(__dirname,"data","users.json");

const SIGNER_MINT_URL = "http://127.0.0.1:8081/mint";
const SIGNER_BUY_URL  = "http://127.0.0.1:8081/buy";


const app = express();


app.use(cors());

app.use(express.json());

app.use(express.urlencoded({
    extended: true,
    limit: "50mb"
}));

app.use("/uploads", express.static(path.join(__dirname, "uploads")));

const tokenRoute = require("./routes/token");
console.log("TOKEN ROUTE LOADED");

app.use("/api", tokenRoute);





app.use(
  express.static(
    path.join(__dirname,"public")
  )
);

app.post("/api/buy-nft", async (req, res) => {

  try {

      const { postId } = req.body;

      if (!postId) {
          return res.status(400).json({
              success: false,
              error: "postId is required"
          });
      }

      // ===========================
      // Cari Post
      // ===========================

      const posts = loadPosts();

      const post = posts.find(
          p => Number(p.id) === Number(postId)
      );

      if (!post) {
          return res.status(404).json({
              success: false,
              error: "Post not found"
          });
      }

      if (!post.nft) {
          return res.status(400).json({
              success: false,
              error: "Post does not contain NFT"
          });
      }

      // ===========================
      // Payload ke Signer
      // ===========================

      const payload = {

          tokenId: post.nft.tokenId,

          name: post.nft.name,

          image: `${BASE_URL}/${post.nft.image}`,

          metadata: JSON.stringify({

              postId: post.id,

              owner: post.owner,

              nft: post.nft

          })

      };

      console.log("SEND TO SIGNER");
      console.log(payload);

      // ===========================
      // Mint NFT
      // ===========================

      const signerRes = await fetch(
          SIGNER_MINT_URL,
          {
              method: "POST",
              headers: {
                  "Content-Type": "application/json"
              },
              body: JSON.stringify(payload)
          }
      );

      const text = await signerRes.text();

      let data;

      try {

          data = JSON.parse(text);

      } catch {

          data = {
              raw: text
          };

      }

      if (!signerRes.ok) {

          return res.status(500).json({
              success: false,
              error: data
          });

      }

      // ===========================
      // Ambil txHash
      // ===========================

      const txHash =
          data.txHash ||
          data.hash ||
          data.rpc?.raw ||
          data.raw ||
          "";

      // ===========================
      // Update posts.json
      // ===========================

      post.nft.bought = true;

      post.nft.txHash = txHash;

      post.nft.mintedAt = Date.now();

      savePosts(posts);

      // ===========================
      // Simpan ke minted-nfts.json
      // ===========================

      const mintedNFTs = loadMintedNFTs();

      const alreadyExist = mintedNFTs.find(
          nft => nft.tokenId === post.nft.tokenId
      );

      if (!alreadyExist) {

          mintedNFTs.push({

              tokenId: post.nft.tokenId,

              txHash: txHash,

              owner: post.owner,

              postId: post.id,

              name: post.nft.name,

              image: post.nft.image,

              price: post.nft.price,

              emoji: post.nft.emoji,

              bg: post.nft.bg || "",

              metadata: post.nft,

              mintedAt: Date.now()

          });

          saveMintedNFTs(mintedNFTs);

          console.log("NFT SAVED TO minted-nfts.json");

      } else {

          console.log("NFT ALREADY EXISTS");

      }

      // ===========================
      // Response
      // ===========================

      res.json({

          success: true,

          signer: data,

          txHash,

          nft: post.nft

      });

  } catch (err) {

      console.error(err);

      res.status(500).json({

          success: false,

          error: err.message

      });

  }

});

app.put("/api/posts/:id/bought", (req, res) => {

  const postId = Number(req.params.id);

  const posts = loadPosts();

  const post = posts.find(p => p.id === postId);

  if (!post) {
      return res.status(404).json({
          success: false,
          message: "Post not found"
      });
  }

  if (post.nft) {
      post.nft.bought = true;
  }

  savePosts(posts);

  res.json({
      success: true
  });

});

app.post("/api/mint-nft", async (req, res) => {
  try {
    const tokenId = "inkarnasi-" + Date.now();

    const payload = {
      tokenId,
      name: req.body.name || "Inkarnasi",
      image: req.body.image || "http://localhost:3001/images/inkarnasi.png",
      metadata: req.body.metadata || "Manga access NFT for Inkarnasi",
    };

    console.log("SEND TO SIGNER:", payload);

    const signerRes = await fetch(SIGNER_MINT_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    const text = await signerRes.text();

    let data;
    try {
      data = JSON.parse(text);
    } catch {
      data = { raw: text };
    }

    if (!signerRes.ok) {
      return res.status(500).json({
        success: false,
        error: data,
      });
    }

    res.json({
      success: true,
      tokenId,
      signer: data,
    });
  } catch (err) {
    console.error("MINT SERVER ERROR:", err);

    res.status(500).json({
      success: false,
      error: err.message,
    });
  }
});



const UPLOAD_DIR =
path.join(
  __dirname,
  "public",
  "uploads"
);



if(!fs.existsSync(UPLOAD_DIR)){
  fs.mkdirSync(
    UPLOAD_DIR,
    { recursive:true }
  );
}

const BASE_URL = `http://localhost:${PORT}`;


const storage =
multer.diskStorage({

  destination:
  function(req,file,cb){

    cb(
      null,
      UPLOAD_DIR
    );

  },

  filename:
  function(req,file,cb){

    const ext =
    path.extname(
      file.originalname
    );

    cb(
      null,
      Date.now() + ext
    );

  }

});

const upload =
multer({
  storage
});

require("dotenv").config();

const CANOPY_EXE =
process.env.CANOPY_EXE;

const CANOPY_PASSWORD =
process.env.CANOPY_PASSWORD;


console.log("CANOPY_EXE:", CANOPY_EXE);
console.log("CANOPY_PASSWORD:", CANOPY_PASSWORD);

/* =========================
   DATABASE
========================= */



function loadDB() {
  if (!fs.existsSync(DB_FILE)) {
    fs.writeFileSync(DB_FILE, "[]");
  }

  return JSON.parse(fs.readFileSync(DB_FILE, "utf8"));
}

function saveDB(data) {
  fs.writeFileSync(DB_FILE, JSON.stringify(data, null, 2));
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


const NFT_DB_FILE =
path.join(__dirname, "data", "minted-nfts.json");

function loadMintedNFTs() {

    if (!fs.existsSync(NFT_DB_FILE)) {
        fs.writeFileSync(NFT_DB_FILE, "[]");
    }

    return JSON.parse(
        fs.readFileSync(
            NFT_DB_FILE,
            "utf8"
        )
    );

}

function saveMintedNFTs(data) {

    fs.writeFileSync(
        NFT_DB_FILE,
        JSON.stringify(data, null, 2)
    );

}

/* =========================
   GET MY NFTS
========================= */

app.get("/api/my-nfts/:owner", (req, res) => {

  try {

      const owner =
          req.params.owner.toLowerCase();

      const mintedNFTs =
          loadMintedNFTs();

      const myNFTs =
          mintedNFTs.filter(nft => {

              if (!nft.owner) {
                  return false;
              }

              return (
                  nft.owner.toLowerCase() === owner
              );

          });

      res.json({

          success: true,

          total: myNFTs.length,

          nfts: myNFTs

      });

  } catch (err) {

      console.error(err);

      res.status(500).json({

          success: false,

          error: err.message

      });

  }

});

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

app.get("/api/posts/:ethAddress", (req, res) => {

  const eth = (req.params.ethAddress || "").toLowerCase();

  const allPosts = loadPosts();

  console.log(allPosts);

  const posts = allPosts.filter(post => {

      if (!post) {
          console.log("POST UNDEFINED");
          return false;
      }

      const owner = (post.owner ?? "").toLowerCase();

      return owner === eth;

  });

  res.json({
      success: true,
      posts
  });

});


/* =========================
   CREATE CANOPY ADDRESS
========================= */

function createCanopyWallet() {

  const nickname =
    "hunter_" + Date.now();

  const cmd =
    `${CANOPY_EXE} admin ks-new-key ` +
    `--nickname ${nickname} ` +
    `--password ${CANOPY_PASSWORD}`;

  console.log("");
  console.log("===== CREATE WALLET =====");
  console.log(cmd);
  console.log("=========================");
  console.log("");

  const output =
  execSync(
    cmd,
    {
      encoding: "utf8",
      shell: "/bin/bash"
    }
  ).trim();

  console.log(output);

  const address =
    output.replace(/"/g, "");

  return {
    nickname,
    address
  };
}
  // send

  function sendFaucetTx(toAddress) {
    try {
  
      const FAUCET_ADDRESS = process.env.CANOPY_SIGNER_ADDRESS;
  
      if (!FAUCET_ADDRESS) {
        throw new Error("CANOPY_SIGNER_ADDRESS is not set");
      }
  
      console.log("FAUCET:", FAUCET_ADDRESS);
      console.log("TO:", toAddress);
  
      const args = [
        "admin",
        "tx-send",
        FAUCET_ADDRESS,
        toAddress,
        "20000",
        "--password",
        CANOPY_PASSWORD
      ];
  
      console.log("EXEC:", CANOPY_EXE, args.join(" "));
  
      const output = execFileSync(
        CANOPY_EXE,
        args,
        {
          encoding: "utf8"
        }
      ).trim();
  
      console.log("OUTPUT:", output);
  
      return {
        success: true,
        txHash: output
      };
  
    } catch (err) {
  
      console.error("MESSAGE:", err.message);
  
      console.error(
        "STDOUT:",
        err.stdout?.toString() || "(empty)"
      );
  
      console.error(
        "STDERR:",
        err.stderr?.toString() || "(empty)"
      );
  
      return {
        success: false,
        error:
          err.stderr?.toString() ||
          err.message
      };
    }
  }

  function sendTreasuryTx(nickname, toAddress, amount) {

    console.log("========== SEND TREASURY ==========");
    console.log("Nickname :", nickname);
    console.log("Receiver :", toAddress);
    console.log("Amount   :", amount);

    try {

        const output = execFileSync(

            CANOPY_EXE,

            [

                "admin",

                "tx-send",

                nickname,

                toAddress,

                String(amount),

                "--password",

                CANOPY_PASSWORD

            ],

            {

                encoding: "utf8"

            }

        ).trim();

        return {

            success: true,

            txHash: output

        };

    }

    catch(err){

        console.error(err);

        return {

            success: false,

            error: err.message

        };

    }

}
  
/* =========================
   LOGIN
========================= */
app.post("/api/login", (req, res) => {

  try {

      const { ethAddress } = req.body;

      if (!ethAddress) {
          return res.status(400).json({
              success:false,
              error:"ethAddress required"
          });
      }

      const users = loadDB();

      let isNewUser = false;

      let user = users.find(
          u =>
              u.ethAddress &&
              u.ethAddress.toLowerCase() ===
              ethAddress.toLowerCase()
      );

      if(!user){

          isNewUser = true;

          const canopyWallet = createCanopyWallet();

          user = {

              ethAddress,

              canopyAddress: canopyWallet.address,

              name:"",
              username:"",
              bio:"",
              photo:"",
              cover:"",

              createdAt:Date.now(),
              updatedAt:Date.now()

          };

          users.push(user);

          saveDB(users);

      }

      const needProfile =

      !user.name ||
      !user.username ||
      !user.photo ||
      !user.cover;
  
  res.json({
  
      success:true,
  
      isNewUser,
  
      needProfile,
  
      user
  
  });

  } catch(err){

      console.error(err);

      res.status(500).json({

          success:false,

          error:err.message

      });

  }

});

app.get("/api/user/:ethAddress", (req, res) => {
  try {
    const eth = req.params.ethAddress.toLowerCase();
    const users = loadDB();
    const user = users.find(u => u.ethAddress.toLowerCase() === eth);

    if(!user){
      return res.status(404).json({ success: false, error: "User not found" });
    }

    res.json({ success: true, user });
  } catch(err) {
    res.status(500).json({ success: false, error: err.message });
  }
});

// fee

app.get("/api/token-fee", async (req, res) => {
  try {

    // sementara hardcode dulu
    res.json({
      success: true,
      createTokenFee: 1000
    });

  } catch (err) {
    res.status(500).json({
      success: false,
      error: err.message
    });
  }
});

   
/* =========================
   GET USER
========================= */
app.get(
  "/api/canopy/:ethAddress",
  async (req,res)=>{

    try{

      const eth =
        req.params.ethAddress
        .toLowerCase();

      const users = loadDB();

      const user =
        users.find(
          u =>
          u.ethAddress.toLowerCase()
          === eth
        );

      if(!user){
        return res.status(404).json({
          success:false
        });
      }

      res.json({
        success:true,
        canopyAddress:
          user.canopyAddress
      });

    }catch(err){

      res.status(500).json({
        success:false,
        error:err.message
      });

    }

});


/* =========================
   WALLET SIGNUP
========================= */

app.post("/api/wallet-signup", (req, res) => {

  try {

    const {
      ethAddress,
      displayName,
      username,
      photo
    } = req.body;

    if (!ethAddress) {
      return res.status(400).json({
        success: false,
        message: "Wallet address required."
      });
    }

    if (!displayName) {
      return res.status(400).json({
        success: false,
        message: "Display Name required."
      });
    }

    if (!username) {
      return res.status(400).json({
        success: false,
        message: "Username required."
      });
    }

    const users = loadDB();

    // cek ETH sudah pernah daftar
    const exist = users.find(
      u =>
        u.ethAddress &&
        u.ethAddress.toLowerCase() ===
        ethAddress.toLowerCase()
    );

    if (exist) {

      return res.json({
        success: false,
        message: "Wallet already registered."
      });

    }

    console.log("=================================");
    console.log("CREATE CANOPY WALLET");
    console.log("=================================");

    // ===============================
    // BUAT WALLET CANOPY
    // ===============================

    const canopyWallet = createCanopyWallet();

    console.log("CANOPY ADDRESS :", canopyWallet.address);

    // ===============================
    // SIMPAN USER
    // ===============================

    const user = {

      ethAddress,

      canopyAddress: canopyWallet.address,

      nickname: canopyWallet.nickname,

      name: displayName,

      username:
        username.startsWith("@")
          ? username
          : "@" + username,

      bio: "",

      photo: photo || "",

      cover: "",

      createdAt: Date.now(),

      updatedAt: Date.now()

    };

    users.push(user);

    saveDB(users);

    console.log("USER SAVED");
    console.log(user);

    // ===============================
    // OPTIONAL FAUCET
    // ===============================

    try {

      const faucet = sendFaucetTx(
        canopyWallet.address
      );

      console.log("FAUCET RESULT:", faucet);

    } catch (err) {

      console.log("Faucet Failed:", err.message);

    }

    res.json({

      success: true,

      user

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


// faucet
app.post(
  "/api/faucet",
  async (req,res)=>{

    try{

      const {
        address
      } = req.body;

      if(!address){

        return res.json({
          success:false,
          error:"Address kosong"
        });

      }

      if(
        !/^[0-9a-fA-F]{40}$/
        .test(address)
      ){

        return res.json({
          success:false,
          error:"Format address tidak valid"
        });

      }

      const result =
        sendFaucetTx(
          address
        );

      if(!result.success){

        return res.json({
          success:false,
          error:result.error
        });

      }

      res.json({
        success:true,
        txHash:
          result.txHash,
        message:
          "1000 CNP berhasil dikirim"
      });

    }
    catch(err){

      console.error("===== EXEC ERROR =====");
    
      console.error("MESSAGE:");
      console.error(err.message);
    
      console.error("STDOUT:");
      console.error(
        err.stdout
          ? err.stdout.toString()
          : "(empty)"
      );
    
      console.error("STDERR:");
      console.error(
        err.stderr
          ? err.stderr.toString()
          : "(empty)"
      );
    
      console.error("======================");
    
      return {
        success:false,
        error:
          err.stderr
            ? err.stderr.toString()
            : err.message
      };
    }

  }
);




/* =========================
   UPDATE PROFILE
========================= */
app.post(
  "/api/upload-avatar",
  upload.single("photo"),
  (req,res)=>{

    try{

      const {
        ethAddress
      } = req.body;

      const users =
      loadDB();

      const user =
      users.find(
        u =>
        u.ethAddress.toLowerCase()
        ===
        ethAddress.toLowerCase()
      );

      if(!user){

        return res.status(404).json({
          success:false,
          error:"User not found"
        });

      }

      user.photo = `${BASE_URL}/uploads/${req.file.filename}`;

      saveDB(users);

      res.json({
        success:true,
        photo:user.photo
      });

    }catch(err){

      console.error(err);

      res.status(500).json({
        success:false
      });

    }

});

// cover


app.post(
  "/api/upload-cover",
  upload.single("cover"),
  (req,res)=>{

    try{

      const {
        ethAddress
      } = req.body;

      const users =
      loadDB();

      const user =
      users.find(
        u =>
        u.ethAddress.toLowerCase()
        ===
        ethAddress.toLowerCase()
      );

      if(!user){

        return res.status(404).json({
          success:false
        });

      }

      user.cover = `${BASE_URL}/uploads/${req.file.filename}`;

      saveDB(users);

      res.json({
        success:true,
        cover:user.cover
      });

    }catch(err){

      console.error(err);

      res.status(500).json({
        success:false
      });

    }

});

// search
app.get("/api/search-users", (req,res)=>{

  const q = (req.query.q || "").toLowerCase();

  const users = loadDB();

  const result = users.filter(u =>
    (u.username || "").toLowerCase().includes(q) ||
    (u.name || "").toLowerCase().includes(q)
  );

  res.json({
    success:true,
    users: result.slice(0,10)
  });
});

// update profile

app.post("/api/update-profile", (req,res)=>{

  try{

    const {
      ethAddress,
      name,
      username,
      bio
  } = req.body;

    if(!ethAddress){
      return res.status(400).json({
        success:false,
        error:"ethAddress required"
      });
    }

    const users = loadDB();

    let user = users.find(
      u =>
      u.ethAddress &&
      u.ethAddress.toLowerCase() ===
      ethAddress.toLowerCase()
    );

    // jika user belum ada -> buat baru
    if(!user){

      user = {
        ethAddress,
        name:"",
        username:"",
        bio:"",
        photo:"",
        cover:"",
        createdAt: Date.now()
      };

      users.push(user);
    }

    user.name = name || user.name || "";

    user.username =
      username
      ? (
          username.startsWith("@")
          ? username
          : "@" + username
        )
      : (user.username || "");

    user.bio = bio || "";

    user.updatedAt = Date.now();

    saveDB(users);

    res.json({
      success:true,
      user
    });

  }catch(err){

    console.error(err);

    res.status(500).json({
      success:false,
      error:err.message
    });

  }

});


app.get("/api/test-wallet", (req, res) => {
  res.json({
    creator: process.env.TEST_ADDRESS,
    publicKey: process.env.TEST_PUBLIC_KEY,
    privateKey: process.env.TEST_PRIVATE_KEY
  });
});

// app.post(
//   "/api/create-token",
//   upload.single("logo"),
//   async (req, res) => {
//     try {



//       const logo =
//         req.file
//           ? `${BASE_URL}/uploads/${req.file.filename}`
//           : "";
//           const {
//             creator,
//             publicKey,
//             privateKey,
//             name,
//             symbol,
//             supply,
//             description
//           } = req.body;
          
//           const signerResp = await fetch(
//             "http://127.0.0.1:8081/create-token",
//             {
//                 method: "POST",
//                 headers: {
//                   "Content-Type":
//                     "application/json"
//                 },
//               body: JSON.stringify({
//                 creator,
//                 publicKey,
//                 privateKey,
//                 name,
//                 symbol,
//                 supply: Number(supply),
//                 logo,
//                 description
//               })
//             }
//           );
          
//           const text =
//             await signerResp.text();
          
//           console.log(
//             "SIGNER STATUS:",
//             signerResp.status
//           );
          
//           console.log(
//             "SIGNER TEXT:",
//             text
//           );
          
//           let result;
//           try {
//             result = JSON.parse(text);
//           } catch (e) {
//             console.error("RAW RESPONSE:", text);
//             throw new Error("Signer bukan JSON valid");
//           }
          
//           res.json(result);

//     } catch (err) {
//       console.error(err);

//       res.status(500).json({
//         success: false,
//         error: err.message
//       });
//     }
//   }
// );

/* =========================
   ALL USERS
========================= */

app.get("/api/users", (req, res) => {

  const db = loadDB();

  res.json({
    success: true,
    users: db
  });

});


// post


app.get("/posts", (req,res)=>{

  res.json(
      loadPosts()
  );

});

app.post("/posts", (req, res) => {

  const posts = loadPosts();

  if (!req.body || Object.keys(req.body).length === 0) {
      return res.status(400).json({
          success: false,
          error: "Empty body"
      });
  }

  posts.unshift(req.body);

  savePosts(posts);

  res.json({
      success: true
  });

});

app.delete("/posts/:id",(req,res)=>{

  const id = Number(req.params.id);

  const posts = loadPosts();

  const filtered = posts.filter(p => p && p.id !== id);

  savePosts(filtered);

  res.json({
      success:true
  });

});

// app.post("/api/mint-nft", async (req, res) => {
//   try {
//     const tokenId = "inkarnasi-" + Date.now();

//     const payload = {
//       tokenId,
//       name: req.body.name || "Inkarnasi",
//       image: req.body.image || "http://localhost:3001/images/inkarnasi.png",
//       metadata: req.body.metadata || "Manga access NFT for Inkarnasi",
//     };

//     console.log("SEND TO SIGNER:", payload);

//     const signerRes = await fetch(SIGNER_MINT_URL, {
//       method: "POST",
//       headers: {
//         "Content-Type": "application/json",
//       },
//       body: JSON.stringify(payload),
//     });

//     const text = await signerRes.text();

//     let data;
//     try {
//       data = JSON.parse(text);
//     } catch {
//       data = { raw: text };
//     }

//     if (!signerRes.ok) {
//       return res.status(500).json({
//         success: false,
//         error: data,
//       });
//     }

//     res.json({
//       success: true,
//       tokenId,
//       signer: data,
//     });
//   } catch (err) {
//     console.error("MINT SERVER ERROR:", err);

//     res.status(500).json({
//       success: false,
//       error: err.message,
//     });
//   }
// });


/* =========================
   SERVER
========================= */

app.listen(PORT, "0.0.0.0", () => {

  console.log("");
  console.log("🚀 HUNTER SERVER");
  console.log(`🌐 http://0.0.0.0:${PORT}`);
  console.log("");

});


