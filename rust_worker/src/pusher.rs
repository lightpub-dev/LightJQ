use derive_builder::Builder;

use crate::models::JobRequestBuilderDefault;

#[derive(Debug, Builder)]
pub struct PusherOptions {
    default_timeout: chrono::Duration,
    default_max_retry: i32,
}

#[derive(Debug)]
pub struct Pusher {
    options: PusherOptions,
}

impl JobRequestBuilderDefault for &Pusher {
    fn default_max_retry(&self) -> i32 {
        self.options.default_max_retry
    }

    fn default_timeout(&self) -> chrono::Duration {
        self.options.default_timeout
    }
}

impl Pusher {
    pub fn new(options: PusherOptions) -> Self {
        Pusher { options }
    }
}
