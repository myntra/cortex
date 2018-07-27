import React, { Component } from 'react';
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

const ScriptCard = (props) => {
  const { classes, script, handleScriptSelection } = props;
  return (
    <Paper className={classes.paper} onClick={() => handleScriptSelection(script.id)}>
      <Typography variant="title" > {script.id} </Typography>
    </Paper>
  )
}

class Sandbox extends Component{
  constructor(props,context){
    super(props,context);
    this.state = {

    }
  }

  render(){
    const { classes, theme } = this.props;
    return(
      <div>
        <Grid container >
          <Button onClick={() => this.setState({ scriptDialogOpen: true })} variant="contained" color="primary" className={classes.button}>
            Add
            <AddIcon className={classes.rightIcon} />
          </Button>
          <Button onClick={this.handleScriptDelete} variant="contained" color="secondary" className={classes.button}>
            Delete
            <DeleteIcon className={classes.rightIcon} />
          </Button>
          {/* {(this.state.scriptID !== "") ?
            <Button onClick={this.handleScriptUpdate} style={{ marginLeft: 'auto' }} variant="contained" color="primary" className={classes.button}>
              Update
              <SubmitIcon className={classes.rightIcon} />
            </Button> : false
          } */}
        </Grid>
        <Grid container spacing={24}>
          <Grid item xs={4}>
            <List>
              {/* {this.state.scriptList.map((script, index) => (
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
              ))} */}
            </List>
          </Grid>
          <Grid item xs={8}>
              <Paper className={classes.paper}>
                {/* <AceEditor
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
                  }} /> */}
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
                    value={``}
                    setOptions={{
                    enableBasicAutocompletion: false,
                    enableLiveAutocompletion: false,
                    enableSnippets: true,
                    showLineNumbers: true,
                    tabSize: 2,
                    }}/>
              </Paper>
          </Grid>
        </Grid>
      </div>
    )
  }

}

export default withStyles(styles, { withTheme: true }) (Sandbox);