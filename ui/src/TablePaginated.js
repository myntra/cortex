import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableFooter from '@material-ui/core/TableFooter';
import TablePagination from '@material-ui/core/TablePagination';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import IconButton from '@material-ui/core/IconButton';
import FirstPageIcon from '@material-ui/icons/FirstPage';
import KeyboardArrowLeft from '@material-ui/icons/KeyboardArrowLeft';
import KeyboardArrowRight from '@material-ui/icons/KeyboardArrowRight';
import LastPageIcon from '@material-ui/icons/LastPage';
import JSONTree from 'react-json-view';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Button from '@material-ui/core/Button';
import OutlineIcon from '@material-ui/icons/OpenInNew';

// shopping_basket

const actionsStyles = theme => ({
  root: {
    flexShrink: 0,
    color: theme.palette.text.secondary,
    marginLeft: theme.spacing.unit * 2.5
  },
});

class TablePaginationActions extends React.Component {
  handleFirstPageButtonClick = event => {
    this.props.onChangePage(event, 0);
  };

  handleBackButtonClick = event => {
    this.props.onChangePage(event, this.props.page - 1);
  };

  handleNextButtonClick = event => {
    this.props.onChangePage(event, this.props.page + 1);
  };

  handleLastPageButtonClick = event => {
    this.props.onChangePage(
      event,
      Math.max(0, Math.ceil(this.props.count / this.props.rowsPerPage) - 1),
    );
  };

  render() {
    const { classes, count, page, rowsPerPage, theme } = this.props;

    return (
      <div className={classes.root}>
        <IconButton
          onClick={this.handleFirstPageButtonClick}
          disabled={page === 0}
          aria-label="First Page"
        >
          {theme.direction === 'rtl' ? <LastPageIcon /> : <FirstPageIcon />}
        </IconButton>
        <IconButton
          onClick={this.handleBackButtonClick}
          disabled={page === 0}
          aria-label="Previous Page"
        >
          {theme.direction === 'rtl' ? <KeyboardArrowRight /> : <KeyboardArrowLeft />}
        </IconButton>
        <IconButton
          onClick={this.handleNextButtonClick}
          disabled={page >= Math.ceil(count / rowsPerPage) - 1}
          aria-label="Next Page"
        >
          {theme.direction === 'rtl' ? <KeyboardArrowLeft /> : <KeyboardArrowRight />}
        </IconButton>
        <IconButton
          onClick={this.handleLastPageButtonClick}
          disabled={page >= Math.ceil(count / rowsPerPage) - 1}
          aria-label="Last Page"
        >
          {theme.direction === 'rtl' ? <FirstPageIcon /> : <LastPageIcon />}
        </IconButton>
      </div>
    );
  }
}

TablePaginationActions.propTypes = {
  classes: PropTypes.object.isRequired,
  count: PropTypes.number.isRequired,
  onChangePage: PropTypes.func.isRequired,
  page: PropTypes.number.isRequired,
  rowsPerPage: PropTypes.number.isRequired,
  theme: PropTypes.object.isRequired,
};

const TablePaginationActionsWrapped = withStyles(actionsStyles, { withTheme: true })(
  TablePaginationActions,
);

const styles = theme => ({
  root: {
    width: '100%',
    marginTop: theme.spacing.unit * 3,
  },
  table: {
    minWidth: 500
  },
  tableWrapper: {
    overflowX: 'auto',
    marginTop:'6px'
  },
});

class TablePaginated extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      data: props.data,
      page: 0,
      rowsPerPage: 5,
      ruleDialogOpen:false,
      selectedHistory:""
    };
  }

  handleChangePage = (event, page) => {
    this.setState({ page });
  };

  handleChangeRowsPerPage = event => {
    this.setState({ rowsPerPage: event.target.value });
  };

  componentWillReceiveProps = (nextProps) => {
    if(nextProps.data !== this.props.data){
      this.setState({data:nextProps.data});
    }
  }

  handleScriptResult = (data) => {
    let finalString;

    if(typeof data === "string"){
      finalString = {
        result : data
      }
    }else if(typeof data === "object" && data !== null){
      finalString = data;
    }else{
      finalString = {
        result : "No result data"
      }
    }
    this.setState({selectedHistory:finalString,
      ruleDialogOpen:true})
  }

  handleBucketResult = (data) => {
    this.setState({selectedHistory:data,
      ruleDialogOpen:true})
  }

  getFormattedDateTime = (utcDateTime) => {
    let d = new Date(utcDateTime);
    let dformat = [d.getMonth()+1, d.getDate(), d.getFullYear()].join('/') + ' ' +
      [d.getHours(), d.getMinutes(), d.getSeconds()].join(':');
    return dformat;
  }

  render() {
    let self = this;
    let colCss = {
      fontSize:'12px',
      fontWeight:300,
      textAlign:'center'
    }
    const { classes } = this.props;
    const { data, rowsPerPage, page } = this.state;
    const emptyRows = rowsPerPage - Math.min(rowsPerPage, data.length - page * rowsPerPage);

    return (
      <Paper className={classes.root} style={{marginTop:'8px'}}>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.ruleDialogOpen}
          onClose={()=> this.setState({ruleDialogOpen:false})}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Bucket Info</DialogTitle>
          <DialogContent>
            <JSONTree src={this.state.selectedHistory} />
          </DialogContent>
          <DialogActions>
            <Button onClick={()=> this.setState({ruleDialogOpen:false})}
              style={{ fontSize: '12px' }}
              color="primary">
              Close
              </Button>
          </DialogActions>
        </Dialog>
        <div className={classes.tableWrapper}>
          <Table className={classes.table}>
            <TableHead>
              <TableRow>
                <TableCell component="th" scope="row" style={{fontSize:'12px',
                  fontWeight:600,
                  textAlign:'center'}}>
                  ID
                </TableCell>
                <TableCell component="th" numeric style={{fontSize:'12px',
                  fontWeight:600,
                  textAlign:'center'}}>
                  Event Bucket
                </TableCell>
                <TableCell component="th" numeric style={{fontSize:'12px',
                  fontWeight:600,
                  textAlign:'center'}}>
                  Script Result
                </TableCell>
                <TableCell component="th" numeric style={{fontSize:'12px',
                  fontWeight:600,
                  textAlign:'center'}}>
                  Hook Status
                </TableCell>
                <TableCell component="th" numeric style={{fontSize:'12px',
                  fontWeight:600,
                  width:'25%',
                  textAlign:'center'}}>
                  Time
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((row,index) => {
                return (
                  <TableRow key={row.id}>
                    <TableCell component="th" scope="row" style={colCss}>
                      {index  + 1}
                    </TableCell>
                    <TableCell numeric style={colCss} onClick={() => self.handleBucketResult(row.bucket)}>
                      <OutlineIcon style={{cursor:'pointer',fontSize: '15px'}} className={classes.rightIcon} />
                    </TableCell>
                    <TableCell numeric style={colCss} onClick={() => self.handleScriptResult(row.script_result)}>
                      <OutlineIcon style={{cursor:'pointer',fontSize: '15px'}} className={classes.rightIcon} />
                    </TableCell>
                    <TableCell numeric style={colCss}>
                      {row.hook_status_code}
                    </TableCell>
                    <TableCell numeric style={colCss}>
                      {self.getFormattedDateTime(row.created_at)}
                    </TableCell>
                  </TableRow>
                );
              })}
              {emptyRows > 0 && (
                <TableRow style={{ height: 48 * emptyRows }}>
                  <TableCell colSpan={6} />
                </TableRow>
              )}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TablePagination
                  colSpan={3}
                  count={data.length}
                  rowsPerPage={rowsPerPage}
                  page={page}
                  onChangePage={this.handleChangePage}
                  onChangeRowsPerPage={this.handleChangeRowsPerPage}
                  ActionsComponent={TablePaginationActionsWrapped}
                />
              </TableRow>
            </TableFooter>
          </Table>
        </div>
      </Paper>
    );
  }
}

TablePaginated.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(TablePaginated);