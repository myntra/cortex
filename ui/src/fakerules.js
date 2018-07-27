const fakerules = [
    {
        id:'1',
        title: "Revenue Down rule",
        scriptID: "revenue.js",
        hook_endpoint: "http://localhost:4000",
        hook_retry: "3",
        event_type_patterns: "com.acme.order.node1.check_disk,com.acme.checkout.node1.check_cpu",
        dwell: "120",
        dwell_deadline: "100",
        max_dwell: "240",
    },
    {
        id:'2',
        title: "Cart Down rule",
        scriptID: "card_down.js",
        hook_endpoint: "http://localhost:4000",
        hook_retry: "3",
        event_type_patterns: "com.acme.cart.node1.check_disk",
        dwell: "120",
        dwell_deadline: "100",
        max_dwell: "240",
    },
    {
        id:'3',
        title: "Style Down rule",
        scriptID: "style_down.js",
        hook_endpoint: "http://localhost:4000",
        hook_retry: "3",
        event_type_patterns: "com.acme.style.node1.check_node",
        dwell: "120",
        dwell_deadline: "100",
        max_dwell: "240",
    }
];

export default fakerules;