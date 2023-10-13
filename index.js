const getFetchable = async () => {
    const res = await fetch('http://localhost:9999/[fetchable] key')
    const data = await res.text();
    console.log("getFetchable", data);
}

const getTimeless = async () => {
    const res = await fetch('http://localhost:9999/[timeless] key')
    const data = await res.text();
    console.log("getTimeless", data);
}

const getRegular = async () => {
    const res = await fetch('http://localhost:9999/[regular] key')
    const data = await res.text();
    console.log("getRegular", data);
}

const getRedeemEstimate = async () => {
    const res = await fetch('http://localhost:9999/REDEEM_ESTIMATE')
    const data = await res.text();
    console.log("getRedeemEstimate", data);
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
    // let counter = 10;
    // const interval = setInterval(() => {
    //     console.log("counter", counter);
    //     getFetchable();
    //     counter--;
    //     if (counter === 0) {
    //         clearInterval(interval);
    //     }
    // }, 1000)

    // setInterval(getTimeless, timer.timeless);
    // setInterval(getRegular, timer.regular);
    let counter = 0;
    setInterval(() => {
        console.log("counter", counter);
        getRedeemEstimate();
        counter++;
    }, timer.redeemEstimate);
}

main();
