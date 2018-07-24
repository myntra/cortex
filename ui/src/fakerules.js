const fakerules = [
    {
        title: "Revenue Down rule",
        scriptID: "revenue.js",
        hook_endpoint: "http://localhost:4000",
        hook_retry: "3",
        event_types: "com.acme.order.node1.check_disk,com.acme.checkout.node1.check_cpu",
        wait_window: "120",
        wait_window_threshold: "100",
        max_wait_window: "240",
    },
    {
        title: "Cart Down rule",
        scriptID: "card_down.js",
        hook_endpoint: "http://localhost:4000",
        hook_retry: "3",
        event_types: "com.acme.cart.node1.check_disk",
        wait_window: "120",
        wait_window_threshold: "100",
        max_wait_window: "240",
    },
    {
        title: "Style Down rule",
        scriptID: "style_down.js",
        hook_endpoint: "http://localhost:4000",
        hook_retry: "3",
        event_types: "com.acme.style.node1.check_node",
        wait_window: "120",
        wait_window_threshold: "100",
        max_wait_window: "240",
    }
];

export default fakerules;