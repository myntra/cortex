import React, { Component } from 'react';
import logo from './logo.svg';
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
import AddIcon from '@material-ui/icons/Add';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import Checkbox from '@material-ui/core/Checkbox';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import withMobileDialog from '@material-ui/core/withMobileDialog';
import brace from 'brace';
import AceEditor from 'react-ace';

import 'brace/mode/javascript';
import 'brace/theme/github';

import fakerules from './fakerules';
import fakescripts from './fakescripts';

const styles = theme => ({
  root: {
    flexGrow: 1,
  },
  paper: {
    padding: theme.spacing.unit * 2,
    textAlign: 'center',
    color: theme.palette.text.secondary,
  },
  heading: {
    fontSize: theme.typography.pxToRem(20),
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
    max_dwell: { type: "string", title: "Maximum Wait Window(seconds)", default: "240" },

  }
};

const uiSchema = {
  event_type_patterns: {
    "ui:widget": "textarea"
  }
};



const log = (type) => console.log.bind(console, type);


const RuleCard = (props) => {
  const { classes, rule } = props;
  return (
    <ExpansionPanel>
      <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
        <Typography className={classes.heading}>{rule && rule.title}</Typography>
      </ExpansionPanelSummary>
      <ExpansionPanelDetails>
          <Form
            uiSchema={uiSchema}
            schema={schema}
            formData={rule}
            onChange={log("changed")}
            onSubmit={log("submitted")}
            onError={log("errors")} />
      </ExpansionPanelDetails>

    </ExpansionPanel>)
}

const ScriptCard = (props) => {
  const { classes, script } = props;
  return (
    <Paper className={classes.paper}>
      <Typography variant="title" > {script.title} </Typography>
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
    rulesChecked: [0],
    ruleDialogOpen: false,
  };

  handleTabChange = (event, tabValue) => {
    this.setState({ tabValue });
  };

  handleChangeTabIndex = index => {
    this.setState({ tabValue: index });
  };

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
  };

  handleRuleDialogOpen = () => {
    this.setState({ ruleDialogOpen: true });
  };

  handleRuleDialogClose = () => {
    this.setState({ ruleDialogOpen: false });
  };

  handleOnScriptChange = (newValue) => {

  }


  render() {
    const { classes, theme } = this.props;
    fakerules.map((rule, index) => console.log(rule, index))

    
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
                onChange={log("changed")}
                onError={log("errors")} > 
                <div>
                  {/* empty div hides default submit button */}
                </div>
              </Form>
          </DialogContent>
          <DialogActions>
            <Button onClick={this.handleRuleDialogClose} color="primary">
              Cancel
            </Button>
            <Button onClick={this.handleRuleDialogClose} color="primary" autoFocus>
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
          centered
        >
          <Tab label="Rules" />
          <Tab label="Scripts" />
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

              <Button variant="contained" color="secondary" className={classes.button}>
                Delete
                <DeleteIcon className={classes.rightIcon} />
              </Button>
            </Grid>
            <Grid container spacing={24}>
              <Grid item xs={6}>
                <List>
                  {fakerules.map((rule, index) => (
                    <Grid key={index} container>
                      <Grid item xs={2}>
                        <ListItem
                          key={index}
                          role={undefined}
                          onClick={this.handleRuleCheckToggle(index)}
                          className={classes.listItem}
                        >
                          <Checkbox
                            checked={this.state.rulesChecked.indexOf(index) !== -1}
                            tabIndex={-1}
                            disableRipple
                          />
                        </ListItem>
                      </Grid>
                      <Grid item xs={10} >
                        <RuleCard key={index} classes={classes} rule={rule} />
                      </Grid>
                    </Grid>
                  ))}
                </List>
              </Grid>
              <Grid item xs>
                {/* <Paper className={classes.paper}></Paper> */}
              </Grid>
            </Grid>
          </TabContainer>
          <TabContainer dir={theme.direction}>
            <Grid container >
                <Button onClick={this.handleRuleDialogOpen} variant="contained" color="primary" className={classes.button}>
                  Add
                <AddIcon className={classes.rightIcon} />
              </Button>

              <Button variant="contained" color="secondary" className={classes.button}>
                Delete
                <DeleteIcon className={classes.rightIcon} />
              </Button>
            </Grid>
            <Grid container spacing={24}>
              <Grid item xs={4}>
                <List>
                  {fakescripts.map((script, index) => (
                    <Grid key={index} container>
                      <Grid item xs={2}>
                        <ListItem
                          key={index}
                          role={undefined}
                          onClick={this.handleRuleCheckToggle(index)}
                          className={classes.listItem}
                        >
                          <Checkbox
                            checked={this.state.rulesChecked.indexOf(index) !== -1}
                            tabIndex={-1}
                            disableRipple
                          />
                        </ListItem>
                      </Grid>
                      <Grid item xs={10} >
                        <ScriptCard key={index} classes={classes} script={script} />
                      </Grid>
                    </Grid>
                  ))}
                </List>
              </Grid>
              <Grid item xs={8}>
                
                <Paper className={classes.paper}>
                <AceEditor
                  mode="javascript"
                  theme="github"
                  name="blah2"
                  onLoad={this.onLoad}
                  onChange={this.onChange}
                  fontSize={14}
                  showPrintMargin={true}
                  showGutter={true}
                  highlightActiveLine={true}
                  value={
                    `function onLoad(editor) {
                        console.log("i've loaded");
                     }`
                  }
                  setOptions={{
                  enableBasicAutocompletion: true,
                  enableLiveAutocompletion: true,
                  enableSnippets: true,
                  showLineNumbers: true,
                  tabSize: 2,
                }}/>
            
                </Paper>
              </Grid>
            </Grid>
          </TabContainer>
        </SwipeableViews>

      </div>
    );
  }
}

export default withStyles(styles, { withTheme: true })(App);
