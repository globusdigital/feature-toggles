#[macro_use(bson, doc)]
extern crate bson;

mod filters;
mod storage;
mod messaging;

use std::env;
use std::net;

use warp::Filter;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(name = "feature-toggles-service")]
struct Opt {
    #[structopt(short = "a", long = "addr", default_value = "127.0.0.1:80")]
    addr: String,

    #[structopt(short = "s", long = "storage-type", default_value = "mem")]
    storage: storage::Type,

    #[structopt(long = "mongo-db", default_value = "mongodb:://mongo/featuretoggles")]
    mongodb: String,

    #[structopt(short = "m", long = "messaging", default_value = "noop")]
    messaging: messaging::Type,

    #[structopt(long = "nats", default_value = "nats://nats:4222")]
    nats: String,
}

#[tokio::main]
async fn main() {
    let opt = Opt::from_args();

    if env::var_os("RUST_LOG").is_none() {
        // Set `RUST_LOG=todos=debug` to see debug logs,
        // this only shows access logs.
        env::set_var("RUST_LOG", "todos=info");
    }

    let addr: net::SocketAddr = opt.addr.parse().ok().expect("Error parsing address");

    let store = storage::MemStore::new();
    let messaging = messaging::Noop::new();

    // GET /hello/warp => 200 OK with body "Hello, warp!"
    let hello = warp::path!("hello" / String)
        .map(|name| format!("Hello, {}!", name));

    println!("Listening on: {}", addr);
    warp::serve(hello)
        .run(addr)
        .await;
}
