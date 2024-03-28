use std::{collections::HashMap, fmt::Debug};

use serde::{Deserialize, Serialize};

use crate::models::{Job, JobResult};

pub type JobName = String;

pub struct Worker {
    handlers: HashMap<JobName, Box<dyn Fn(Job<rmpv::Value>) -> JobResult>>,
}

impl Debug for Worker {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Worker")
            .field("handlers", &self.handlers.keys().collect::<Vec<_>>())
            .finish()
    }
}

impl Worker {
    pub fn new() -> Self {
        Worker {
            handlers: HashMap::new(),
        }
    }

    pub fn register_handler<A, F, R>(&mut self, name: JobName, handler: F)
    where
        A: for<'de> Deserialize<'de>,
        F: Fn(A) -> R + 'static,
        R: Into<JobResult>,
    {
        self.handlers.insert(
            name,
            Box::new(move |job| {
                let argument = job.get_argument();
                let argument = rmpv::ext::from_value::<A>(argument).unwrap(); // TODO: error handling
                handler(argument).into()
            }),
        );
    }
}

#[derive(Deserialize)]
struct Sample1 {
    a: String,
}

#[derive(Deserialize)]
struct Sample2 {
    b: i32,
}

fn handler1(data: Sample1) -> JobResult {
    todo!()
}

fn handler2(data: Sample2) -> JobResult {
    todo!()
}
