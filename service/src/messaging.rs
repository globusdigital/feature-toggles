use std::error::Error;
use rants::{Client};
use super::storage;

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

pub trait Messaging {
    fn send(&mut self, flags: Vec<storage::Flag>) -> Result<(), Box<dyn Error>>;
}

pub struct Noop {}

impl Messaging for Noop {
    fn send(&mut self, _: Vec<storage::Flag>) -> Result<(), Box<dyn Error>> {
        Ok(())
    }
}

impl Noop {
    pub fn new() -> Box<dyn Messaging> {
        Box::new(Noop{})
    }
}

pub struct Nats {
    client: Client,
}

impl Messaging for Nats {
    fn send(&mut self, _: Vec<storage::Flag>) -> Result<(), Box<dyn Error>> {
        Ok(())
    }
}

impl Nats {
    pub fn new(addr: &str) -> Result<Box<dyn Messaging>, Box<dyn Error>> {
        let client = rants::Client::new(vec![addr.parse()?]);
        Ok(Box::new(Nats{client}))
    }
}
