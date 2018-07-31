import React, { Component } from 'react';
import './App.css';
import Form from "react-jsonschema-form";
import CssBaseline from '@material-ui/core/CssBaseline';
import { withStyles } from '@material-ui/core/styles';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import SwipeableViews from 'react-swipeable-views';
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import Button from '@material-ui/core/Button';
import DeleteIcon from '@material-ui/icons/Delete';
import SubmitIcon from '@material-ui/icons/Save';
import AddIcon from '@material-ui/icons/Add';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import Checkbox from '@material-ui/core/Checkbox';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import AceEditor from 'react-ace';
import JSONTree from 'react-json-tree'
import TablePaginated from './TablePaginated';
import base64 from 'base-64';

import 'brace/mode/javascript';
import 'brace/theme/github';
import Sandbox from './Sandbox';

const styles = theme => ({
  root: {
    flexGrow: 1,
  },
  paper: {
    padding: theme.spacing.unit * 2,
    textAlign: 'center',
    color: theme.palette.text.secondary,
    cursor: 'pointer'
  },
  heading: {
    fontSize: '12px',
    fontWeight: theme.typography.fontWeightRegular,
  },
  button: {
    margin: theme.spacing.unit,
  },
  leftIcon: {
    marginRight: theme.spacing.unit,
  },
  rightIcon: {
    marginLeft: theme.spacing.unit,
  },
  iconSmall: {
    fontSize: 20,
  },
});

const schema = {
  type: "object",
  title: "",
  script_id: "",
  hook_endpoint: "",
  hook_retry: "",
  event_type_patterns: "",
  dwell: "",
  dwell_deadline: "",
  max_dwell: "",
  required: ["title", "event_type_patterns"],
  properties: {
    title: { type: "string", title: "Title", default: "A new rule" },
    script_id: { type: "string", title: "Script", default: "default.js" },
    hook_endpoint: { type: "string", title: "Hook Endpoint", default: "http://localhost:4000" },
    hook_retry: { type: "number", title: "Hook Retry", default: 2 },
    event_type_patterns: { type: "string", title: "Match Event Types", default: "com.acme.node1.cpu,com.apple.node2.cpu" },
    dwell: { type: "number", title: "Wait Window(seconds)", default: 120 },
    dwell_deadline: { type: "number", title: "Wait Window Threshold(seconds)", default: 100 },
    max_dwell: { type: "number", title: "Maximum Wait Window(seconds)", default: 240 }
  }
}

const scriptSchema = {
  type: "object",
  id: "",
  required: ["id"],
  properties: {
    id: { type: "string", title: "Script ID", default: "" }
  }
};

const uiSchema = {
  event_type_patterns: {
    "ui:widget": "textarea"
  }
};

const uiSchemaScript = {
  Data: {
    "ui:widget": "textarea"
  },
  "ui:widget": "string",
  "ui:help": "Hint: like default.js"
};

const log = (type) => console.log.bind(console, type);

const RuleCard = (props) => {
  const { classes, rule, handleChangeEvent, handleSubmitEvent, handlePannelExpansion } = props;
  let updatedRule = rule;
  updatedRule.event_type_patterns = rule.event_type_patterns.toString().split().join(",")
  return (
    <ExpansionPanel onChange={(event, flag) => handlePannelExpansion(event, flag, rule.id)}>
      <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
        <Typography className={classes.heading}>{rule && rule.title}</Typography>
      </ExpansionPanelSummary>
      <ExpansionPanelDetails>
        <div style={{ width: '100%', fontSize: '12px' }}>
          <Form
            uiSchema={uiSchema}
            schema={schema}
            formData={updatedRule}
            onChange={(event) => handleChangeEvent(event, rule.id)}
            onSubmit={(event) => handleSubmitEvent(event, rule.id)}
            onError={log("errors")} />
        </div>
      </ExpansionPanelDetails>
    </ExpansionPanel>
  )
}

const ScriptCard = (props) => {
  const { classes, script, handleScriptSelection } = props;
  return (
    <Paper className={classes.paper} onClick={() => handleScriptSelection(script)}>
      <Typography variant="title" > {script} </Typography>
    </Paper>
  )
}

function TabContainer({ children, dir }) {
  return (
    <Typography component="div" dir={dir} style={{ padding: 8 * 3 }}>
      {children}
    </Typography>
  );
}

class App extends Component {

  state = {
    tabValue: 0,
    rulesChecked: [],
    scriptsChecked: [],
    ruleDialogOpen: false,
    scriptDialogOpen: false,
    ruleList: [],
    scriptIDList:[],
    scriptList: [],
    newRule: {},
    newScript: {},
    expansionFlag: false,
    scriptText: "",
    scriptID: "",
    history:""
  }

  componentDidMount() {
    this.fetchRules();
    this.fetchScripts();
  }

  fetchRules = () => {
    let self = this;
    fetch('/rules')
      .then(function (response) {
        if (response.ok) {
          return response.json();
        }
        else {
          throw new Error('Something went wrong. Unable to fetch list of rules')
        }
      })
      .then(function (data) {
        self.setState({ ruleList: data })
      })
      .catch((error) => {
        alert(error);
      })
  }

  fetchScripts = () => {
    let self = this;
    fetch('/scripts')
      .then(function (response) {
        if (response.ok) {
          return response.json();
        } else {
          throw new Error('Something went wrong. Unable to fetch list of scripts');
        }
      })
      .then(function (data) {
        self.setState({ scriptIDList: data })
      })
      .catch((error) => {
        alert(error);
      })
  }

  fetchHistory = (flag, id) => {
    let self = this;
    let url = "/rules/" + id + "/executions" 
    fetch(url)
      .then(function (response) {
        if (response.ok) {
          return response.json();
        }
        else {
          throw new Error('Something went wrong. Unable to fetch list of rules')
        }
      })
      .then(function (data) {
        self.setState({
          selectedID: id,
          expansionFlag: flag,
          history:data
        })
      })
      .catch((error) => {
        alert(error);
      })
  }

  getBytesFromString = (str) => {
    let bytes = new Uint8Array(str.length);
    for (let i = 0; i < bytes.length; i++) {
      bytes[i] = str.charCodeAt(i);
    }
    return bytes;
  }

  getStringFromBytes = (byteArray) => {
    var str = ""
    for (let i = 0; i < byteArray.length; i++) {
      str = str + String.fromCharCode(byteArray[i]);
    }
    return str;
  }

  handleTabChange = (event, tabValue) => {
    this.setState({ tabValue });
  };

  handleChangeTabIndex = index => {
    this.setState({ tabValue: index });
  }

  handleRuleCheckToggle = value => () => {
    const { rulesChecked } = this.state;
    const currentIndex = rulesChecked.indexOf(value);
    const newChecked = [...rulesChecked];

    if (currentIndex === -1) {
      newChecked.push(value);
    } else {
      newChecked.splice(currentIndex, 1);
    }

    this.setState({
      rulesChecked: newChecked,
    });
  }

  handleScriptCheckToggle = (value) => {
    const { scriptsChecked } = this.state;
    const currentIndex = scriptsChecked.indexOf(value);
    const newChecked = [...scriptsChecked];

    if (currentIndex === -1) {
      newChecked.push(value);
    } else {
      newChecked.splice(currentIndex, 1);
    }
    this.setState({
      scriptsChecked: newChecked,
    });
  }

  handleRuleDialogOpen = () => {
    this.setState({ ruleDialogOpen: true });
  }

  handleRuleDialogClose = () => {
    this.setState({ ruleDialogOpen: false });
  }

  handleRuleDialogSave = () => {
    let self = this;
    const { newRule } = this.state
    if(Object.keys(newRule).length === 0){
      alert("Default rule cannot be used. Sample data is just for reference");
      return
    }
    let eventPatterns = [];
    try {
      eventPatterns = newRule.event_type_patterns.split(",");
    } catch (error) {
      alert("Unable to create string array list for events pattern. Please check event patterns");
      return
    }
    let json = {
      "title": newRule.title,
      "script_id": newRule.script_id,
      "hook_endpoint": newRule.hook_endpoint,
      "hook_retry": parseInt(newRule.hook_retry),
      "event_type_patterns": eventPatterns,
      "dwell": parseInt(newRule.dwell),
      "dwell_deadline": parseInt(newRule.dwell_deadline),
      "max_dwell": parseInt(newRule.max_dwell)
    }
    fetch('/rules', {
      method: "POST",
      body: JSON.stringify(json)
    })
      .then(function (response) {
        self.setState({ ruleDialogOpen: false });
        if (response.ok) {
          return response.json();
        } else {
          throw new Error('Something went wrong. Unable to create new rule');
        }
      })
      .then(function (data) {
        console.log("Updated successfully", data);
        self.fetchRules()
      })
      .catch((error) => {
        alert(error);
      });
  }

  handleScriptDialogSave = () => {
    let self = this;
    const { newScript } = this.state
    let defaultFunc =
    `import http from "k6/http";
    // Reference: https://docs.k6.io/docs/http-requests
    let result = null;
    export default function(bucket) {
        console.log(bucket) 
    }`;
    let obj = {
      id : newScript.id,
      Data : base64.encode(defaultFunc)
    }
    console.log('Save script', obj);
    fetch('/scripts', {
      method: "POST",
      body: JSON.stringify(obj)
    })
      .then(function (response) {
        if (response.ok) {
          console.log("Updated successfully");
          self.fetchScripts()
          self.setState({ scriptDialogOpen: false });
        } else {
          throw new Error('Something went wrong. Unable to create new script');
        }
      })
      .catch((error) => {
        alert(error);
      });
  }

  handleRuleChange = (event, id) => {
    let obj = {}
    for (let x in event.formData) {
      obj[x] = event.formData[x]
    }
    let newList = this.state.ruleList.map((item, i) => {
      if (item.id === id) {
        item = obj;
      }
      return item
    })
    this.setState({ ruleList: newList })
  }

  handleRuleSubmit = (event, id) => {
    let self = this;
    const { ruleList } = this.state
    let obj;
    ruleList.forEach((item) => {
      if (item.id === id) {
        obj = item
      }
    })
    if (!obj) {
      console.log("JSON udefined", id, ruleList)
      alert("Found no data to update rule")
      return;
    }
    let eventPatterns = [];
    try {
      eventPatterns = obj.event_type_patterns.split(",");
    } catch (error) {
      alert("Unable to create string array list for events pattern. Please check event patterns");
      return
    }
    let json = {
      "id":obj.id,
      "title": obj.title,
      "script_id": obj.script_id,
      "hook_endpoint": obj.hook_endpoint,
      "hook_retry": parseInt(obj.hook_retry),
      "event_type_patterns": eventPatterns,
      "dwell": parseInt(obj.dwell),
      "dwell_deadline": parseInt(obj.dwell_deadline),
      "max_dwell": parseInt(obj.max_dwell)
    }
    fetch('/rules', {
      method: "PUT",
      body: JSON.stringify(json)
    })
      .then(function (response) {
        if (response.ok) {
          self.fetchRules();
          console.log("Updated successfully");
        } else {
          throw new Error('Something went wrong. Unable to update rule content');
        }
      })
      .catch((error) => {
        alert(error);
      });
  }

  handleRuleDelete = () => {
    let self = this;
    const { rulesChecked } = this.state
    if (rulesChecked.length < 1) {
      alert("No rule selected for deletion");
    }
    rulesChecked.forEach((item) => {
      let url = "/rules/" + item
      fetch(url, {
        method: "DELETE"
      })
        .then(function (response) {
          if (response.ok) {
            console.log("Updated successfully");
            self.fetchRules()
          } else {
            throw new Error('Something went wrong. Unable to delete rule ' + item);
          }
        })
        .catch((error) => {
          alert(error);
        })
    })
  }

  handleScriptChange = (text) => {
    this.setState({
      scriptText: text
    })
  }

  handleScriptClick = (id) => {
    if(id === ""){
      alert("No id found");
      return;
    }
    let self = this;
    let url = "/scripts/" + id;
    fetch(url)
      .then(function (response) {
        if (response.ok) {
          return response.json();
        } else {
          throw new Error('Something went wrong. Unable to fetch list of scripts');
        }
      })
      .then(function (script) {
        self.setState({ scriptID: script.id,
            scriptText: base64.decode(script.data) })
      })
      .catch((error) => {
        alert(error);
      })
  }

  handleScriptDelete = () => {
    let self = this;
    const { scriptsChecked } = this.state
    if (scriptsChecked.length < 1) {
      alert("No script selected for deletion");
    }
    scriptsChecked.forEach((item) => {
      let url = "/scripts/" + item
      fetch(url, {
        method: "DELETE"
      })
        .then(function (response) {
          if (response.ok) {
            console.log("Updated successfully");
            self.fetchScripts()
          } else {
            throw new Error('Something went wrong. Unable to delete script ' + item);
          }
        })
        .catch((error) => {
          alert(error);
        })
    })
  }

  handleScriptUpdate = () => {
    // TODO: Update not working
    const { scriptID, scriptText } = this.state
    let json = {
      id: scriptID,
      Data: base64.encode(scriptText)
    }
    if (!json) {
      console.log("JSON udefined", scriptID, scriptText)
      return;
    }
    fetch('/scripts', {
      method: "PUT",
      body: JSON.stringify(json)
    })
      .then(function (response) {
        if (response.ok) {
          alert("Updated successfully");
        } else {
          throw new Error('Something went wrong. Unable to update script content');
        }
      })
      .catch((error) => {
        alert(error);
      })
  }

  render() {
    const { classes, theme } = this.props;
    const { history } = this.state;
    // this.state.ruleList.map((rule, index) => console.log(rule, index))
    return (
      <div className={classes.root}>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.ruleDialogOpen}
          onClose={this.handleRuleDialogClose}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Create New Rule</DialogTitle>
          <DialogContent>
            <Form
              uiSchema={uiSchema}
              schema={schema}
              formData={this.state.newRule}
              onChange={(event) => this.setState({ newRule: event.formData })}
              onError={log("errors")} >
              <button type="submit" className="hidden">Submit</button>
            </Form>
          </DialogContent>
          <DialogActions>
            <Button onClick={this.handleRuleDialogClose}
              style={{ fontSize: '12px' }}
              color="primary">
              Cancel
              </Button>
            <Button onClick={this.handleRuleDialogSave}
              style={{ fontSize: '12px' }}
              color="primary" autoFocus>
              Save
              </Button>
          </DialogActions>
        </Dialog>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.scriptDialogOpen}
          onClose={() => this.setState({ scriptDialogOpen: false })}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Create New Script</DialogTitle>
          <DialogContent>
            <Form
              uiSchema={uiSchemaScript}
              schema={scriptSchema}
              formData={this.state.newScript}
              onChange={(event) => this.setState({ newScript: event.formData })}
              onError={log("errors")} >
              <button type="submit" className="hidden">Submit</button>
            </Form>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => this.setState({ scriptDialogOpen: false })} color="primary">
              Cancel
              </Button>
            <Button onClick={this.handleScriptDialogSave} color="primary" autoFocus>
              Save
              </Button>
          </DialogActions>
        </Dialog>
        <CssBaseline />
        <AppBar position="static" color="primary">
          <Toolbar>
            <Typography variant="display2" color="inherit">
              Cortex
          </Typography>
          </Toolbar>
        </AppBar>
        <Tabs
          value={this.state.tabValue}
          indicatorColor="primary"
          textColor="primary"
          onChange={this.handleTabChange}
          style={{ fontSize: '24px' }}
          centered
        >
          <Tab label="Rules" />
          <Tab label="Scripts" />
          <Tab label="Playground" />
        </Tabs>
        <SwipeableViews
          axis={theme.direction === 'rtl' ? 'x-reverse' : 'x'}
          index={this.state.tabValue}
          onChangeIndex={this.handleChangeTabIndex}
        >
          <TabContainer dir={theme.direction}>
            <Grid container >
              <Button onClick={this.handleRuleDialogOpen} variant="contained" color="primary" className={classes.button}>
                Add
                <AddIcon className={classes.rightIcon} />
              </Button>

              <Button variant="contained" color="secondary" onClick={this.handleRuleDelete} className={classes.button}>
                Delete
                <DeleteIcon className={classes.rightIcon} />
              </Button>
            </Grid>
            <Grid container spacing={24}>
              <Grid item xs={6}>
                <List>
                  {this.state.ruleList.map((rule, index) => (
                    <Grid key={index} container>
                      <Grid item xs={2}>
                        <ListItem
                          key={index}
                          role={undefined}
                          onClick={this.handleRuleCheckToggle(rule.id)}
                          className={classes.listItem}
                        >
                          <Checkbox
                            checked={this.state.rulesChecked.indexOf(rule.id) !== -1}
                            tabIndex={-1}
                            disableRipple
                          />
                        </ListItem>
                      </Grid>
                      <Grid item xs={10} >
                        <RuleCard key={index}
                          handlePannelExpansion={(event, flag, id) => this.fetchHistory(flag, id)}
                          handleChangeEvent={(event, id) => this.handleRuleChange(event, id)}
                          handleSubmitEvent={(event, id) => this.handleRuleSubmit(event, id)}
                          classes={classes} rule={rule} />
                      </Grid>
                    </Grid>
                  ))}
                </List>
              </Grid>
              <Grid item xs>
                {
                  (this.state.expansionFlag) ?
                    <div style={{fontSize:"12px"}}>
                      {/* <TablePaginated /> */}
                      <h1>Execution History</h1>
                      <JSONTree data={history} />
                    </div>
                    : false
                }
              </Grid>
            </Grid>
          </TabContainer>
          <TabContainer dir={theme.direction}>
            <Grid container >
              <Button onClick={() => this.setState({ scriptDialogOpen: true })} variant="contained" color="primary" className={classes.button}>
                Add
                <AddIcon className={classes.rightIcon} />
              </Button>
              <Button onClick={this.handleScriptDelete} variant="contained" color="secondary" className={classes.button}>
                Delete
                <DeleteIcon className={classes.rightIcon} />
              </Button>
              {(this.state.scriptID !== "") ?
                <Button onClick={this.handleScriptUpdate} style={{ marginLeft: 'auto' }} variant="contained" color="primary" className={classes.button}>
                  Update
                  <SubmitIcon className={classes.rightIcon} />
                </Button> : false
              }
            </Grid>
            <Grid container spacing={24}>
              <Grid item xs={4}>
                <List>
                  {this.state.scriptIDList.map((script, index) => (
                    <Grid key={index} container>
                      <Grid item xs={2}>
                        <ListItem
                          key={index}
                          role={undefined}
                          onClick={() => this.handleScriptCheckToggle(script)}
                          className={classes.listItem}
                        >
                          <Checkbox
                            checked={this.state.scriptsChecked.indexOf(script) !== -1}
                            tabIndex={-1}
                            disableRipple
                          />
                        </ListItem>
                      </Grid>
                      <Grid item xs={10} >
                        <ScriptCard handleScriptSelection={this.handleScriptClick}
                          key={index} classes={classes} script={script} />
                      </Grid>
                    </Grid>
                  ))}
                </List>
              </Grid>
              <Grid item xs={8}>
                {(this.state.scriptID !== "") ?
                  <Paper className={classes.paper}>
                    <AceEditor
                      mode="javascript"
                      theme="github"
                      name="blah2"
                      onLoad={this.onLoad}
                      onChange={(text, event) => this.handleScriptChange(text)}
                      fontSize={14}
                      showPrintMargin={true}
                      style={{width:'100%'}}
                      showGutter={true}
                      highlightActiveLine={true}
                      value={this.state.scriptText}
                      setOptions={{
                        enableBasicAutocompletion: true,
                        enableLiveAutocompletion: true,
                        enableSnippets: true,
                        showLineNumbers: true,
                        tabSize: 2,
                      }} />
                  </Paper> : false
                }
              </Grid>
            </Grid>
          </TabContainer>
          <TabContainer dir={theme.direction}>
            <Sandbox rules={this.state.ruleList} />
          </TabContainer>
        </SwipeableViews>

      </div>
    );
  }
}

export default withStyles(styles, { withTheme: true })(App);