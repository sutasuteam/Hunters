require("dotenv").config();

const { execSync, execFileSync } = require("child_process");


const CANOPY_EXE = process.env.CANOPY_EXE;
const CANOPY_PASSWORD = process.env.CANOPY_PASSWORD;

function createCanopyWallet() {

    const nickname = "hunter_" + Date.now();

    const cmd =
        `${CANOPY_EXE} admin ks-new-key ` +
        `--nickname ${nickname} ` +
        `--password ${CANOPY_PASSWORD}`;

    const output = execSync(cmd, {
        encoding: "utf8",
        shell: "/bin/bash"
    }).trim();

    return {
        nickname,
        address: output.replace(/"/g, "")
    };
}
function sendFaucetTx(toAddress) {

    const faucet = process.env.CANOPY_SIGNER_ADDRESS;

    const output = execFileSync(
        CANOPY_EXE,
        [
            "admin",
            "tx-send",
            faucet,
            toAddress,
            "20000",
            "--password",
            CANOPY_PASSWORD
        ],
        {
            encoding: "utf8"
        }
    ).trim();

    console.log("TREASURY FUNDED:", output);

    return {
        success: true,
        txHash: output
    };
}
function getAddressBalance(address) {

    console.log("QUERY ADDRESS:", address);

    const output = execSync(
        `${CANOPY_EXE} query account ${address}`,
        {
            encoding: "utf8",
            shell: "/bin/bash"
        }
    );

    console.log("QUERY RESULT:");
    console.log(output);

    return JSON.parse(output);

}

function sendTreasuryTx(nickname, toAddress, amount) {

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

        console.log("TREASURY TX:");
        console.log(output);

        return {

            success: true,

            txHash: output

        };

    }

    catch(err){

        console.error("TREASURY ERROR:");
        console.error(err);

        return {

            success: false,

            error: err.message

        };

    }

}

function sendLikePostTx(
    liker,
    owner,
    postId,
    liked
) {

    try {

        const output = execFileSync(

            CANOPY_EXE,

            [

                "admin",

                "tx-like-post",

                liker,

                owner,

                String(postId),

                liked ? "true" : "false",

                "--password",

                CANOPY_PASSWORD

            ],

            {

                encoding: "utf8"

            }

        ).trim();

        console.log("LIKE POST TX:");
        console.log(output);

        return {

            success: true,

            txHash: output

        };

    }

    catch(err){

        console.error("LIKE POST ERROR:");
        console.error(err);

        return {

            success: false,

            error: err.message

        };

    }

}


module.exports = {
    createCanopyWallet,
    sendFaucetTx,
    sendTreasuryTx,
    sendLikePostTx,
    getAddressBalance
};