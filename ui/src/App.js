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
import TablePaginated from './TablePaginated';

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
  scriptID: "",
  hook_endpoint: "",
  hook_retry: "",
  event_type_patterns: "",
  dwell: "",
  dwell_deadline: "",
  max_dwell: "",
  required: ["title", "event_type_patterns"],
  properties: {
    title: { type: "string", title: "Title", default: "A new rule" },
    scriptID: { type: "string", title: "Script", default: "default.js" },
    hook_endpoint: { type: "string", title: "Hook Endpoint", default: "http://localhost:4000" },
    hook_retry: { type: "string", title: "Hook Retry", default: "2" },
    event_type_patterns: { type: "string", title: "Match Event Types", default: "com.acme.node1.cpu,com.apple.node2.cpu" },
    dwell: { type: "string", title: "Wait Window(seconds)", default: "120" },
    dwell_deadline: { type: "string", title: "Wait Window Threshold(seconds)", default: "100" },
    max_dwell: { type: "string", title: "Maximum Wait Window(seconds)", default: "240" }
  }
};

const scriptSchema = {
  type: "object",
  scriptID: "",
  required: ["scriptID"],
  properties: {
    scriptID: { type: "string", title: "Script", default: "" }
  }
};

const uiSchema = {
  event_type_patterns: {
    "ui:widget": "textarea"
  }
};

const uiSchemaScript = {
  "ui:widget": "string",
  "ui:help": "Hint: like default.js"
};

const log = (type) => console.log.bind(console, type);

const RuleCard = (props) => {
  const { classes, rule, handleChangeEvent, handleSubmitEvent, handlePannelExpansion } = props;
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
            formData={rule}
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
    <Paper className={classes.paper} onClick={() => handleScriptSelection(script.id)}>
      <Typography variant="title" > {script.id} </Typography>
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
    tabValue: 2,
    rulesChecked: [],
    scriptsChecked: [],
    ruleDialogOpen: false,
    scriptDialogOpen: false,
    ruleList: [],
    scriptList: [],
    newRule: {},
    newScript: {},
    expansionFlag: false,
    scriptText: "",
    scriptID: ""
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
        self.setState({ scriptList: data })
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
    fetch('/rules', {
      method: "POST",
      body: JSON.stringify(newRule)
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
    console.log('Save script')
    let self = this;
    const { newScript } = this.state
    fetch('/scripts', {
      method: "POST",
      body: JSON.stringify(newScript)
    })
      .then(function (response) {
        if (response.ok) {
          self.setState({ scriptDialogOpen: false });
          return response.json();
        } else {
          throw new Error('Something went wrong. Unable to create new script');
        }
      })
      .then(function (data) {
        console.log("Updated successfully", data);
        self.fetchScripts()
      })
      .catch((error) => {
        alert(error);
      });
  }

  handleRuleExpansion = (event, flag, id) => {
    this.setState({
      selectedID: id,
      expansionFlag: flag
    })
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
    const { ruleList } = this.state
    let json;
    ruleList.forEach((item) => {
      if (item.id === id) {
        json = item
      }
    })
    if (!json) {
      console.log("JSON udefined", id, ruleList)
      alert("Found no data to update rule")
      return;
    }
    console.log("Rule Update", json)
    fetch('/rules', {
      method: "PUT",
      body: JSON.stringify(json)
    })
    .then(function (response) {
      if (response.ok) {
        return response.json();
      } else {
        throw new Error('Something went wrong. Unable to update rule content');
      }
    })
    .then(function (data) {
      console.log("Updated successfully", data);
    })
    .catch((error) => {
      alert(error);
    });
  }

  handleRuleDelete = () => {
    console.log('rulesChecked', this.state.rulesChecked)
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
          return response.json();
        } else {
          throw new Error('Something went wrong. Unable to delete rule ' + item);
        }
      })
      .then(function (data) {
        console.log("Updated successfully", data);
        self.fetchRules()
      })
      .catch((error) => {
        alert(error);
      })
    })
  }

  handleScriptChange = (text, event) => {
    var self = this;
    let scriptItems = this.state.scriptList.map((item, i) => {
      if (item.id === self.state.scriptID) {
        item.Data = self.getBytesFromString(text);
      }
      return item
    })
    this.setState({
      scriptText: text,
      scriptList: scriptItems
    })
  }

  handleScriptClick = (id) => {
    let script;
    this.state.scriptList.forEach((item) => {
      if (item.id === id) {
        script = item;
      }
    })
    let text = this.getStringFromBytes(script.Data)
    this.setState({
      scriptID: script.id,
      scriptText: text
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
          return response.json();
        } else {
          throw new Error('Something went wrong. Unable to delete script ' + item);
        }
      })
      .then(function (data) {
        console.log("Updated successfully", data);
        self.fetchScripts()
      })
      .catch((error) => {
        alert(error);
      })
    })
  }

  handleScriptUpdate = () => {
    const { scriptID, scriptList } = this.state
    let json;
    scriptList.forEach((item) => {
      if (item.id === scriptID) {
        json = item
      }
    })
    if (!json) {
      console.log("JSON udefined", scriptID, scriptList)
      return;
    }
    fetch('/scripts/revenue.js', {
      method: "PUT",
      body: JSON.stringify(json)
    })
    .then(function (response) {
      if (response.ok) {
        return response.json();
      } else {
        throw new Error('Something went wrong. Unable to update script content');
      }
    })
    .then(function (data) {
      console.log("Updated successfully", data);
    })
    .catch((error) => {
      alert(error);
    })
  }

  render() {
    const { classes, theme } = this.props;
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
              style={{fontSize:'12px'}} 
              color="primary">
              Cancel
              </Button>
            <Button onClick={this.handleRuleDialogSave}
              style={{fontSize:'12px'}}
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
          style={{ fontSize: '24px'}}
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
                          handlePannelExpansion={(event, flag, id) => this.handleRuleExpansion(event, flag, id)}
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
                    <TablePaginated />
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
                  {this.state.scriptList.map((script, index) => (
                    <Grid key={index} container>
                      <Grid item xs={2}>
                        <ListItem
                          key={index}
                          role={undefined}
                          onClick={() => this.handleScriptCheckToggle(script.id)}
                          className={classes.listItem}
                        >
                          <Checkbox
                            checked={this.state.scriptsChecked.indexOf(script.id) !== -1}
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
                      onChange={(text, event) => this.handleScriptChange(text, event)}
                      fontSize={14}
                      showPrintMargin={true}
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