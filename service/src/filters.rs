use std::convert::Infallible;

use serde::Deserialize;
use warp::Filter;

use super::storage::Store;
use super::messaging::Bus;

#[derive(Debug, Deserialize)]
pub struct ListOptions {
    pub service_name: Option<String>,
}

pub fn flags(
    store: impl Store, bus: impl Bus,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    flags_list(store.clone())
}

fn flags_list(
    store: impl Store,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("flags").and(warp::get()).and(with_store(store)).and(warp::query::<ListOptions>()).and_then(flags_list_handler)
}

fn with_store(store: impl Store) -> impl Filter<Extract = (impl Store,), Error = std::convert::Infallible> + Clone {
    warp::any().map(|| store.clone())
}

async fn flags_list_handler(store: impl Store, opts: ListOptions) -> Result<impl warp::Reply, Infallible>{
    // Just return a JSON array of todos, applying the limit and offset.
    let flags = store.get_flags(opts.service_name);
    Ok(warp::reply::json(&flags.map_err(|err| err.to_string())))
}