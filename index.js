const getRedeemEstimate = async () => {
    const res = await fetch('http://localhost:9999/REDEEM_ESTIMATE')
    const data = await res.text();
    return `getRedeemEstimate ${data}`
}

const randomCancel = (fn) => {
    return () => {
        const rand = Math.random()
        if (rand > 0.5) {
            fn();
        }
    }
}

const timer = {
    fetchable: 20000,
    timeless: 5000,
    regular: 1000,
    redeemEstimate: 1000,
}

const main = async () => {
    let counter = 0;
    setInterval(async () => {
        console.log("counter", counter);
        counter++;

        const startTime = performance.now();
        const data = await getRedeemEstimate();
        const endTime = performance.now(); 
        console.log("data", data, "elapsed", endTime - startTime, "ms");
    }, timer.redeemEstimate);
}

main();
