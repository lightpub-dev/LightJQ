use crate::models::{Job, JobRequest, JobResult, Result, Worker};
use redis::Commands;
use serde::{Deserialize, Serialize};

#[derive(Debug)]
pub struct RedisTransport {
    r: redis::Client,
}

const WORKER_REGISTER_QUEUE: &str = "jq:workerRegister";
const GLOBAL_QUEUE: &str = "jq:globalQueue";
const RESULT_QUEUE: &str = "jq:resultQueue";
const JOB_REGISTER_QUEUE: &str = "jq:jobRegister";

impl RedisTransport {
    pub fn new(r: redis::Client) -> Self {
        RedisTransport { r }
    }

    pub async fn register_worker(&mut self, worker: &Worker) -> Result<()> {
        self.r
            .rpush(WORKER_REGISTER_QUEUE, worker)
            .map_err(Into::into)
    }

    pub async fn blocking_pop_job<'de, A: Deserialize<'de>>(&mut self) -> Result<Job<A>> {
        self.r.blpop(GLOBAL_QUEUE, 0.0).map_err(Into::into)
    }

    pub async fn report_job_result(&mut self, result: &JobResult) -> Result<()> {
        self.r.rpush(RESULT_QUEUE, result).map_err(Into::into)
    }

    pub async fn enqueue_job<A: Serialize>(&mut self, job: &JobRequest<A>) -> Result<()> {
        self.r.rpush(JOB_REGISTER_QUEUE, job).map_err(Into::into)
    }
}
