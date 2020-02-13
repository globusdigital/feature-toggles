use std::error::Error;
use futures::executor::block_on;
use rants::{Client};
use super::storage;

const NATS_SUBJECT: &'static str = "feature-toggles";

#[derive(Debug)]
pub enum Type {
    Noop,
    NATS
}

impl std::str::FromStr for Type {
    type Err = &'static str;
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "noop" => Ok(Type::Noop),
            "nats" => Ok(Type::NATS),
            _ => Err("Error parsing messaging type"),
        }
    }
}

pub trait Bus {
    fn send(&self, flags: Vec<storage::Flag>) -> Result<(), Box<dyn Error>>;
}

pub struct Noop {}

impl Bus for Noop {
    fn send(&self, _: Vec<storage::Flag>) -> Result<(), Box<dyn Error>> {
        Ok(())
    }
}

impl Noop {
    pub fn new() -> Box<dyn Bus> {
        Box::new(Noop{})
    }
}

pub struct Nats {
    client: Client,
}

impl Bus for Nats {
    fn send(&self, flags: Vec<storage::Flag>) -> Result<(), Box<dyn Error>> {
        let payload = serde_json::to_string(&flags)?;
        let subject = NATS_SUBJECT.parse::<rants::Subject>().unwrap();
        block_on(self.client.publish(&subject, payload.as_bytes()))?;
        Ok(())
    }
}

impl Nats {
    pub async fn new(addr: &str) -> Result<Box<dyn Bus>, Box<dyn Error>> {
        let client = rants::Client::new(vec![addr.parse()?]);
        client.connect().await;
        Ok(Box::new(Nats{client}))
    }
}
