import React, { Component } from 'react';
import Form from "react-jsonschema-form";
import CssBaseline from '@material-ui/core/CssBaseline';
import { withStyles } from '@material-ui/core/styles';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import DeleteIcon from '@material-ui/icons/Delete';
import SubmitIcon from '@material-ui/icons/Done';
import AddIcon from '@material-ui/icons/Add';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import Checkbox from '@material-ui/core/Checkbox';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import AceEditor from 'react-ace';
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import uuidv4 from "uuid/v4";

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


const eventSchema = {
  type: "object",
  eventID: "",
  delay: "",
  source: "",
  eventTime:"",
  data: "",
  eventType: "",
  required: ["eventID", "delay", "data", "source", "eventType","eventTime"],
  properties: {
    delay: { type: "string", title: "Submit Delay", default: "1" },
    eventType: { type: "string", title: "Event Type", default: "acme.prod.site247.search_down" },
    source: { type: "string", title: "Event Source", default: "site247" },
    data: { type: "string", title: "Event Data", default: '{"appinfoA":"abc","appinfoB": 123,"appinfoC": true}' },
    eventID: { type: "string", title: "Event ID", default: uuidv4() },
    eventTime: { type: "string", title: "Event Time", default: "2018-04-05T17:31:00Z" }
  }
};

const eventSchemaCreate = {
  type: "object",
  delay: "",
  source: "",
  eventTime:"",
  data: "",
  eventType: "",
  required: ["eventID", "delay", "data", "source", "eventType","eventTime"],
  properties: {
    delay: { type: "string", title: "Submit Delay", default: "1" },
    eventType: { type: "string", title: "Event Type", default: "acme.prod.site247.search_down" },
    source: { type: "string", title: "Event Source", default: "site247" },
    data: { type: "string", title: "Event Data", default: '{"appinfoA":"abc","appinfoB": 123,"appinfoC": true}' },
    eventTime: { type: "string", title: "Event Time", default: "2018-04-05T17:31:00Z" }
  }
};

const uiSchemaEvent = {
  data: {
    "ui:widget": "textarea"
  },
  delay: {
    "ui:help": "Hint: Delay is for submission delay and it's not part of cloudevents.io spec"
  },
  "ui:widget": "string",
  "ui:help": "Hint: Create event with delay in seconds associated with it"
};

const EventCard = (props) => {
  const { classes, event, handleChangeEvent, handlePannelExpansion,handleSubmitEvent } = props;
  return (
    <ExpansionPanel onChange={(e, flag) => handlePannelExpansion(e, flag, event.eventID)}>
      <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
        <Typography className={classes.heading}>{event && event.eventType}</Typography>
      </ExpansionPanelSummary>
      <ExpansionPanelDetails>
        <div style={{ width: '100%', fontSize: '12px' }}>
          <Form
            uiSchema={uiSchemaEvent}
            schema={eventSchema}
            formData={event}
            onSubmit={(e) => handleSubmitEvent(e, event.eventID)}
            onChange={(e) => handleChangeEvent(e, event.eventID)}
            onError={log("errors")}>
            <button type="submit" className="hidden">Submit</button>
          </Form>
        </div>
      </ExpansionPanelDetails>
    </ExpansionPanel>
  )
}

const log = (type) => console.log.bind(console, type);

class Sandbox extends Component {
  constructor(props, context) {
    super(props, context);
    this.state = {
      events: [
      ],
      newEvent: "",
      eventData: "",
      ruleDialogOpen: false,
      eventsChecked: [],
      selectedEvent: '',
      expansionFlag: false,
      result:[],
      resultFlag:false
    }
  }

  handleEventDialogSave = () => {
    let { events, newEvent } = this.state;
    let flag = false;
    let condition = false;
    if(newEvent === ""){
      alert("Event cannot be created with all default values");
      return
    }
    for(let x in Object.keys(newEvent)){
      if(newEvent[x] === ""){
        alert(x + " cannot be empty");
        condition = true;
        return;
      }
    }
    if(condition){
      return;
    }
    let data;
    try {
      data = JSON.parse(newEvent.data);
    }
    catch(e) {
      alert("Incorrect JSON data");
      return
    }
    events.forEach((item) => {
      if (item.eventID === newEvent.eventID) {
        flag = true;
      }
    })
    if(flag){
      alert("Event name already exists");
      return
    }
    newEvent["eventID"] = uuidv4();
    events.push(newEvent);
    this.setState({
      events: events,
      newEvent: {},
      ruleDialogOpen: false
    })
  }

  handleEventsCheckToggle = (value) => {
    const { eventsChecked } = this.state;
    const currentIndex = eventsChecked.indexOf(value);
    const newChecked = [...eventsChecked];

    if (currentIndex === -1) {
      newChecked.push(value);
    } else {
      newChecked.splice(currentIndex, 1);
    }

    this.setState({
      eventsChecked: newChecked,
    });
  }

  handleEventChange = (value) => {
    const { selectedEvent, events } = this.state;
    let obj;
    let eventList = events.map((item, i) => {
      if (item.eventID === selectedEvent.eventID) {
        item.data = value;
        obj = item;
        return obj;
      } else {
        return item;
      }
    })
    this.setState({
      selectedEvent: obj,
      events: eventList
    })
  }

  handleEventsExecution = () => {
    let self = this;
    const { eventsChecked,events,result } = this.state;
    let eventsList = [];
    if (eventsChecked.length < 1) {
      alert("No event selected for execution");
      return
    }
    events.forEach((item) => {
      if(eventsChecked.indexOf(item.eventID) !== -1){
        let data;
        try {
          data = JSON.parse(item.data);
        }
        catch(e) {
          alert("Incorrect JSON data for event ID " + item.eventID);
          return
        }
        eventsList.push(item);
      }
    })
    this.setState({resultFlag:true})
    eventsList.forEach((event)=>{
      let json = self.cloudEventFormat(event);
      let time = parseInt(event.delay) * 1000;
      setTimeout(function(jsonData) {
        // API calls for updating events
        fetch('/events', {
          method: "POST",
          body: JSON.stringify(jsonData)
        })
        .then(function (response) {
          if (response.ok) {
            return response.json();
          } else {
            throw new Error('Something went wrong. Unable to push event');
          }
        })
        .then(function (data) {
          let resultSet = result;
          resultSet.push({"id":jsonData.eventID,"output":" Successfully called API with delay of " + time/1000 + " secs"});
          self.setState({
            result:resultSet
          })
          console.log("Updated successfully", data);
        })
        .catch((error) => {
          let resultSet = result;
          resultSet.push({"id":jsonData.eventID,"output":error});
          self.setState({
            result:resultSet
          })
        });
      }, time, json);
    })
  }

  handleEventExpansion = (event, flag, id) => {
    const { events } = this.state;
    let obj;
    events.forEach((item) => {
      if (item.eventID === id) {
        obj = item;
      }
    })
    if (!obj) {
      alert("No event found");
      return
    }
    this.setState({
      selectedEvent: obj,
      expansionFlag: flag
    })
  }

  cloudEventFormat = (obj) =>{
    let eventDummy = {
      "cloudEventsVersion": "0.1",
      "eventType": obj.eventType,
      "source": obj.source,
      "eventID": obj.eventID,
      "eventTime": obj.eventTime,
      "extensions": {
        "comExampleExtension": "value"
      },
      "contentType": "application/json",
      "data": JSON.parse(obj.data)
    }
    return eventDummy;
  }

  handleOldEventChange = (event, id) => {
    const {events} = this.state
    let obj = event.formData;
    let newList = events.map((item, i) => {
      if (item.eventID === id) {
        item = obj;
      }
      return item
    })
    this.setState({ events: newList, selectedEvent: obj })
  }

  handleEventChangeSubmit = (event, id) => {
    console.log("Events",event,id)
  }

  handleEventDelete = () => {
    const {eventsChecked,events} = this.state
    if(eventsChecked.length < 1){
      alert("No event for deletion");
      return
    }
    let updatedEvents = [];
    events.forEach((item) => {
      if(eventsChecked.indexOf(item.eventID) !== -1){
        console.log('Item',item);
      }
      else{
        updatedEvents.push(item);
      }
    })
    this.setState({events:updatedEvents,eventsChecked:[],selectedEvent:""});
  }

  render() {
    const { classes, theme, rules } = this.props;
    let self = this;
    return (
      <div>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.ruleDialogOpen}
          onClose={() => this.setState({ ruleDialogOpen: false })}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Create New Event</DialogTitle>
          <DialogContent>
            <Form
              uiSchema={uiSchemaEvent}
              schema={eventSchemaCreate}
              formData={this.state.newEvent}
              onChange={(event) => this.setState({ newEvent: event.formData })}
              onError={(error) => console.log("errors", error)} >
              <button type="submit" className="hidden">Submit</button>
            </Form>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => this.setState({ ruleDialogOpen: false })}
              style={{fontSize:'12px'}} 
              color="primary">
              Cancel
              </Button>
            <Button onClick={this.handleEventDialogSave}
              style={{fontSize:'12px'}} 
              color="primary" autoFocus>
              Save
              </Button>
          </DialogActions>
        </Dialog>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.resultFlag}
          onClose={() => this.setState({ resultFlag: false })}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Event Execution Results</DialogTitle>
          <DialogContent>
            {
              this.state.result.map((item) => (
                <div key={item.id}>
                  <span style={{fontWeight:'700',marginRight:'5px'}}>EventID</span> {item.id}
                  <span style={{fontWeight:'700',marginRight:'5px',marginLeft:'5px'}}>Result</span> {item.output}
                </div>
              ))
            }
          </DialogContent>
          <DialogActions>
            <Button onClick={() => this.setState({ resultFlag: false })}
              style={{fontSize:'12px'}} 
              color="primary">
              Close
            </Button>
          </DialogActions>
        </Dialog>
        <Grid container >
          <Button onClick={() => this.setState({ ruleDialogOpen: true })}
            variant="contained"
            color="primary"
            className={classes.button}>
            Add
            <AddIcon className={classes.rightIcon} />
          </Button>
          <Button onClick={this.handleEventDelete}
            variant="contained"
            color="secondary"
            className={classes.button}>
            Delete
            <DeleteIcon className={classes.rightIcon} />
          </Button>
          <Button onClick={this.handleEventsExecution}
            style={{ marginLeft: 'auto' }}
            variant="contained"
            color="primary"
            className={classes.button}>
            Execute
            <SubmitIcon className={classes.rightIcon} />
          </Button>
        </Grid>
        <Grid container spacing={24}>
          <Grid item xs={4}>
            <List>
              {this.state.events.map((event, index) => (
                <Grid key={index} container>
                  <Grid item xs={2}>
                    <ListItem
                      key={index}
                      role={undefined}
                      onClick={() => self.handleEventsCheckToggle(event.eventID)}
                      className={classes.listItem}
                    >
                      <Checkbox
                        checked={self.state.eventsChecked.indexOf(event.eventID) !== -1}
                        tabIndex={-1}
                        disableRipple
                      />
                    </ListItem>
                  </Grid>
                  <Grid item xs={10} >
                    <EventCard key={index}
                      handlePannelExpansion={self.handleEventExpansion}
                      handleSubmitEvent={(event, id) => self.handleEventChangeSubmit(event,id)}
                      handleChangeEvent={(event, id) => self.handleOldEventChange(event, id)}
                      classes={classes} event={event} />
                  </Grid>
                </Grid>
              ))}
            </List>
          </Grid>
          <Grid item xs={8}>
            {(this.state.selectedEvent !== '' && this.state.expansionFlag) ?
              <Paper className={classes.paper}>
                <AceEditor
                  mode="javascript"
                  theme="github"
                  name="blah2"
                  readOnly={true}
                  fontSize={14}
                  showPrintMargin={true}
                  showGutter={true}
                  highlightActiveLine={true}
                  value={`${JSON.stringify(self.cloudEventFormat(self.state.selectedEvent),null, "\t")}`}
                  setOptions={{
                    enableBasicAutocompletion: false,
                    enableLiveAutocompletion: false,
                    enableSnippets: true,
                    showLineNumbers: true,
                    tabSize: 2,
                  }} />
              </Paper> : false
            }
          </Grid>
        </Grid>
      </div>
    )
  }

}

export default withStyles(styles, { withTheme: true })(Sandbox);