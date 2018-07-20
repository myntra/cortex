var ruleRoutes = require("./handlers/ruleRoutes");

function routes(app, handlers) {
    app.get('/', handlers.eventHandler.index);
    app.post('/:ruleType',
        function (req, res) {
            if (ruleRoutes.ruleRoutes.hasOwnProperty(req.params.ruleType)) {
                result = ruleRoutes.ruleRoutes[req.params.ruleType](req)
                return res.json({
                    rca_object: result,
                    status: 0
                })
            } else {
                console.log("Couldnot find the ruleType", req.params.ruleType)
                result = ruleRoutes.ruleRoutes["default"](req)
               return res.json({
                    rca_object: result,
                    status: 1
                })
            }

        });
}

exports.routes = routes;