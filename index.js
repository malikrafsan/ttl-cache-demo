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
    regular: 2500,
}

const main = async () => {
    let counter = 5;
    const interval = setInterval(() => {
        console.log("counter", counter);
        getFetchable();
        counter--;
        if (counter === 0) {
            clearInterval(interval);
        }
    }, 1000)

    // setInterval(getTimeless, timer.timeless);
    setInterval(randomCancel(getRegular), timer.regular);
}

main();
